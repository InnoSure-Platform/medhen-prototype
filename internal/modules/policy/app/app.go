// Package app holds the policy use cases: create quote (prices via rating,
// validates the party) and bind (underwrites, then persists quote + policy +
// outbox event atomically).
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	partyports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/domain"
	ratingports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/ports"
	uwports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/underwriting/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
)

// ErrNotFound is returned when a quote or policy does not exist.
var ErrNotFound = errors.New("policy: not found")

// ErrPartyNotFound is returned when the quoted party does not exist.
var ErrPartyNotFound = errors.New("policy: party not found")

// QuoteRepository persists quotes (uses the ambient UoW connection).
type QuoteRepository interface {
	Save(ctx context.Context, q *domain.Quote) error
	Get(ctx context.Context, tenantID, id string) (*domain.Quote, error)
}

// PolicyRepository persists policies and issues gap-free policy sequences.
type PolicyRepository interface {
	Save(ctx context.Context, p *domain.Policy) error
	Get(ctx context.Context, tenantID, id string) (*domain.Policy, error)
	NextSequence(ctx context.Context, name string) (int64, error)
}

// Deps are the policy service's collaborators, injected at composition time.
type Deps struct {
	DB           *database.DB
	Quotes       QuoteRepository
	Policies     PolicyRepository
	Rating       ratingports.Calculator
	Party        partyports.Reader
	Underwriting uwports.Decider
	Insurer      string // policy-number insurer code, e.g. "EIC"
}

// Service implements the policy use cases.
type Service struct{ deps Deps }

// NewService builds the service.
func NewService(deps Deps) *Service {
	if deps.Insurer == "" {
		deps.Insurer = "EIC"
	}
	return &Service{deps: deps}
}

// CreateQuoteInput is the command payload for pricing a quote.
type CreateQuoteInput struct {
	TenantID       string            `json:"tenant_id"`
	PartyID        string            `json:"party_id"`
	ProductCode    string            `json:"product_code"`
	Coverages      []string          `json:"coverages"`
	RiskDimensions map[string]string `json:"risk_dimensions"`
}

// CreateQuote validates the party, prices the risk via the rating module, and
// persists a QUOTED quote. This is the real in-process rating call that replaces
// the pre-refactor hardcoded 500.00 gRPC stub.
func (s *Service) CreateQuote(ctx context.Context, in CreateQuoteInput) (*domain.Quote, error) {
	if _, err := s.deps.Party.GetParty(ctx, in.TenantID, in.PartyID); err != nil {
		return nil, ErrPartyNotFound
	}

	bd, err := s.deps.Rating.Calculate(ctx, ratingports.PremiumRequest{
		TenantID: in.TenantID, ProductCode: in.ProductCode, AsOfDate: time.Now().UTC(),
		RiskDimensions: in.RiskDimensions, Coverages: in.Coverages,
	})
	if err != nil {
		return nil, fmt.Errorf("policy: rating failed: %w", err)
	}

	q := domain.NewQuote(in.TenantID, in.PartyID, in.ProductCode, in.Coverages, in.RiskDimensions,
		bd.NetPremium, bd.TotalTaxes, bd.GrossPremium, bd.CalculationID)

	if err := s.deps.DB.WithinTx(ctx, func(ctx context.Context) error {
		return s.deps.Quotes.Save(ctx, q)
	}); err != nil {
		return nil, err
	}
	return q, nil
}

// BindQuote underwrites the quote and, on straight-through acceptance, issues a
// policy. The quote transition, the policy insert and the PolicyIssued outbox
// event all commit in ONE transaction (atomic issuance — fixes the pre-refactor
// two-transaction bind that could orphan a BOUND quote).
func (s *Service) BindQuote(ctx context.Context, tenantID, quoteID string) (*domain.Policy, error) {
	q, err := s.deps.Quotes.Get(ctx, tenantID, quoteID)
	if err != nil {
		return nil, err
	}

	decision, err := s.deps.Underwriting.Decide(ctx, uwports.DecisionRequest{
		TenantID: tenantID, ProductCode: q.ProductCode,
		GrossPremium: q.GrossPremium, RiskDimensions: q.RiskDimensions,
	})
	if err != nil {
		return nil, fmt.Errorf("policy: underwriting failed: %w", err)
	}
	switch decision.Outcome {
	case uwports.OutcomeDecline:
		return nil, domain.ErrDeclined
	case uwports.OutcomeRefer:
		return nil, domain.ErrReferred
	}

	var policy *domain.Policy
	err = s.deps.DB.WithinTx(ctx, func(ctx context.Context) error {
		if err := q.Bind(); err != nil {
			return err
		}

		now := time.Now().UTC()
		seqName := fmt.Sprintf("%s-%d", q.ProductCode, now.Year())
		seq, err := s.deps.Policies.NextSequence(ctx, seqName)
		if err != nil {
			return err
		}
		policyNumber := ids.PolicyNumber(s.deps.Insurer, q.ProductCode, now.Year(), seq)
		policy = domain.NewPolicy(policyNumber, tenantID, q.ID, q.PartyID, q.ProductCode, q.GrossPremium, now)

		if err := s.deps.Quotes.Save(ctx, q); err != nil {
			return err
		}
		if err := s.deps.Policies.Save(ctx, policy); err != nil {
			return err
		}

		evt := domain.PolicyIssued{
			PolicyID: policy.ID, PolicyNumber: policy.PolicyNumber, TenantID: tenantID,
			PartyID: policy.PartyID, ProductCode: policy.ProductCode,
			GrossMinor: policy.GrossPremium.Minor(), OccurredAt: now,
		}
		payload, err := json.Marshal(evt)
		if err != nil {
			return fmt.Errorf("policy: marshal event: %w", err)
		}
		return outbox.Write(ctx, s.deps.DB.Conn(ctx), outbox.Message{
			ID: ids.New(), Topic: domain.TopicPolicyIssued,
			AggregateType: "policy", AggregateID: policy.ID, Payload: payload,
		})
	})
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// GetQuote loads a quote.
func (s *Service) GetQuote(ctx context.Context, tenantID, id string) (*domain.Quote, error) {
	return s.deps.Quotes.Get(ctx, tenantID, id)
}

// GetPolicy loads a policy.
func (s *Service) GetPolicy(ctx context.Context, tenantID, id string) (*domain.Policy, error) {
	return s.deps.Policies.Get(ctx, tenantID, id)
}
