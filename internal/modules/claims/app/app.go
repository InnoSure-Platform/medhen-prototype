// Package app holds the claims use cases: file FNOL (validates the policy via the
// policy module's Reader) and fast-track settle. Both persist their aggregate and
// outbox event in one Unit-of-Work.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/domain"
	policyports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
)

var (
	// ErrNotFound is returned when a claim does not exist.
	ErrNotFound = errors.New("claims: not found")
	// ErrPolicyNotFound is returned when the claimed policy does not exist.
	ErrPolicyNotFound = errors.New("claims: policy not found")
	// ErrPolicyNotActive is returned when the policy is not in force.
	ErrPolicyNotActive = errors.New("claims: policy is not in force")
)

// ClaimRepository persists claims (via the ambient UoW connection).
type ClaimRepository interface {
	Save(ctx context.Context, c *domain.Claim) error
	Get(ctx context.Context, tenantID, id string) (*domain.Claim, error)
}

// Deps are the claims service's collaborators.
type Deps struct {
	DB             *database.DB
	Claims         ClaimRepository
	Policy         policyports.Reader
	FastTrackLimit money.Amount
}

// Service implements the claims use cases.
type Service struct{ deps Deps }

// NewService builds the service.
func NewService(deps Deps) *Service { return &Service{deps: deps} }

// FNOLInput is the command payload for reporting a loss.
type FNOLInput struct {
	TenantID     string  `json:"tenant_id"`
	PolicyID     string  `json:"policy_id"`
	Description  string  `json:"description"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	ReserveMinor int64   `json:"reserve_minor"`
}

// FileFNOL validates the policy is in force, records the claim with an initial
// reserve, and emits ClaimFiled — atomically.
func (s *Service) FileFNOL(ctx context.Context, in FNOLInput) (*domain.Claim, error) {
	pv, err := s.deps.Policy.GetPolicy(ctx, in.TenantID, in.PolicyID)
	if err != nil {
		return nil, ErrPolicyNotFound
	}
	if pv.Status != "ISSUED" {
		return nil, ErrPolicyNotActive
	}

	claim, err := domain.NewClaim(in.TenantID, in.PolicyID, pv.PartyID, in.Description,
		in.Latitude, in.Longitude, money.FromMinor(in.ReserveMinor))
	if err != nil {
		return nil, err
	}

	err = s.deps.DB.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.deps.Claims.Save(ctx, claim); err != nil {
			return err
		}
		evt := domain.ClaimFiled{
			ClaimID: claim.ID, TenantID: claim.TenantID, PolicyID: claim.PolicyID,
			PartyID: claim.PartyID, ReserveMinor: claim.Reserve.Minor(), OccurredAt: time.Now().UTC(),
		}
		return writeEvent(ctx, s.deps.DB, domain.TopicClaimFiled, "claim", claim.ID, evt)
	})
	if err != nil {
		return nil, err
	}
	return claim, nil
}

// FastTrackSettle settles a claim within fast-track authority and emits
// ClaimSettled — atomically. Amounts above authority are referred.
func (s *Service) FastTrackSettle(ctx context.Context, tenantID, claimID string, amount money.Amount) (*domain.Claim, error) {
	claim, err := s.deps.Claims.Get(ctx, tenantID, claimID)
	if err != nil {
		return nil, err
	}

	err = s.deps.DB.WithinTx(ctx, func(ctx context.Context) error {
		if err := claim.Settle(amount, s.deps.FastTrackLimit); err != nil {
			return err
		}
		if err := s.deps.Claims.Save(ctx, claim); err != nil {
			return err
		}
		evt := domain.ClaimSettled{
			ClaimID: claim.ID, TenantID: claim.TenantID, PolicyID: claim.PolicyID,
			PartyID: claim.PartyID, AmountMinor: amount.Minor(), OccurredAt: time.Now().UTC(),
		}
		return writeEvent(ctx, s.deps.DB, domain.TopicClaimSettled, "claim", claim.ID, evt)
	})
	if err != nil {
		return nil, err
	}
	return claim, nil
}

// GetClaim loads a claim.
func (s *Service) GetClaim(ctx context.Context, tenantID, id string) (*domain.Claim, error) {
	return s.deps.Claims.Get(ctx, tenantID, id)
}

func writeEvent(ctx context.Context, db *database.DB, topic, aggType, aggID string, evt any) error {
	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("claims: marshal event: %w", err)
	}
	return outbox.Write(ctx, db.Conn(ctx), outbox.Message{
		ID: ids.New(), Topic: topic, AggregateType: aggType, AggregateID: aggID, Payload: payload,
	})
}
