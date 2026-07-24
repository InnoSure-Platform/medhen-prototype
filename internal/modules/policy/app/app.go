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
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
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
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Quote, error)
}

// PolicyRepository persists policies and issues gap-free policy sequences.
type PolicyRepository interface {
	Save(ctx context.Context, p *domain.Policy) error
	Get(ctx context.Context, tenantID, id string) (*domain.Policy, error)
	List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Policy, error)
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

// ListPolicies returns a tenant's issued policies (newest first), paginated.
func (s *Service) ListPolicies(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Policy, error) {
	return s.deps.Policies.List(ctx, tenantID, limit, offset)
}

// ListQuotes returns a tenant's quotes (newest first), paginated.
func (s *Service) ListQuotes(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Quote, error) {
	return s.deps.Quotes.List(ctx, tenantID, limit, offset)
}

// EndorsePolicy adjusts an in-force policy's premium by a signed delta and emits
// PolicyEndorsed — atomically.
func (s *Service) EndorsePolicy(ctx context.Context, tenantID, policyID string, deltaMinor int64, reason string) (*domain.Policy, error) {
	p, err := s.deps.Policies.Get(ctx, tenantID, policyID)
	if err != nil {
		return nil, err
	}
	err = s.deps.DB.WithinTx(ctx, func(ctx context.Context) error {
		if err := p.Endorse(money.FromMinor(deltaMinor)); err != nil {
			return err
		}
		if err := s.deps.Policies.Save(ctx, p); err != nil {
			return err
		}
		return s.writeEvent(ctx, domain.TopicPolicyEndorsed, p.ID, domain.PolicyEndorsed{
			PolicyID: p.ID, TenantID: tenantID, DeltaMinor: deltaMinor,
			NewGrossMinor: p.GrossPremium.Minor(), Reason: reason, OccurredAt: time.Now().UTC(),
		})
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

// CancelPolicy cancels an in-force policy and emits PolicyCancelled (with the
// pro-rata refund estimate) — atomically.
func (s *Service) CancelPolicy(ctx context.Context, tenantID, policyID, reason string) (*domain.Policy, money.Amount, error) {
	p, err := s.deps.Policies.Get(ctx, tenantID, policyID)
	if err != nil {
		return nil, money.Zero(), err
	}
	var refund money.Amount
	err = s.deps.DB.WithinTx(ctx, func(ctx context.Context) error {
		var cerr error
		refund, cerr = p.Cancel(reason, time.Now().UTC())
		if cerr != nil {
			return cerr
		}
		if err := s.deps.Policies.Save(ctx, p); err != nil {
			return err
		}
		return s.writeEvent(ctx, domain.TopicPolicyCancelled, p.ID, domain.PolicyCancelled{
			PolicyID: p.ID, TenantID: tenantID, Reason: reason,
			RefundMinor: refund.Minor(), OccurredAt: time.Now().UTC(),
		})
	})
	if err != nil {
		return nil, money.Zero(), err
	}
	return p, refund, nil
}

// RenewPolicy issues a successor policy for the next term. The renewal is treated
// downstream as a new issuance (emits PolicyIssued so billing/COI/notification
// fire) plus PolicyRenewed for the audit trail — all in ONE transaction.
func (s *Service) RenewPolicy(ctx context.Context, tenantID, policyID string) (*domain.Policy, error) {
	prior, err := s.deps.Policies.Get(ctx, tenantID, policyID)
	if err != nil {
		return nil, err
	}
	if prior.Status != domain.StatusIssued {
		return nil, domain.ErrNotInForce
	}

	var next *domain.Policy
	err = s.deps.DB.WithinTx(ctx, func(ctx context.Context) error {
		now := time.Now().UTC()
		seqName := fmt.Sprintf("%s-%d", prior.ProductCode, prior.EffectiveTo.Year())
		seq, err := s.deps.Policies.NextSequence(ctx, seqName)
		if err != nil {
			return err
		}
		number := ids.PolicyNumber(s.deps.Insurer, prior.ProductCode, prior.EffectiveTo.Year(), seq)
		next = prior.Renew(number)
		if err := s.deps.Policies.Save(ctx, next); err != nil {
			return err
		}
		if err := s.writeEvent(ctx, domain.TopicPolicyRenewed, next.ID, domain.PolicyRenewed{
			PolicyID: next.ID, PriorPolicyID: prior.ID, PolicyNumber: next.PolicyNumber,
			TenantID: tenantID, OccurredAt: now,
		}); err != nil {
			return err
		}
		return s.writeEvent(ctx, domain.TopicPolicyIssued, next.ID, domain.PolicyIssued{
			PolicyID: next.ID, PolicyNumber: next.PolicyNumber, TenantID: tenantID,
			PartyID: next.PartyID, ProductCode: next.ProductCode,
			GrossMinor: next.GrossPremium.Minor(), OccurredAt: now,
		})
	})
	if err != nil {
		return nil, err
	}
	return next, nil
}

func (s *Service) writeEvent(ctx context.Context, topic, aggID string, evt any) error {
	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("policy: marshal event: %w", err)
	}
	return outbox.Write(ctx, s.deps.DB.Conn(ctx), outbox.Message{
		ID: ids.New(), Topic: topic, AggregateType: "policy", AggregateID: aggID, Payload: payload,
	})
}
