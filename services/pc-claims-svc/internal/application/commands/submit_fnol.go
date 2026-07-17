package commands

import (
	"context"
	"time"

	"medhen/pc-claims-svc/internal/domain"
)

type SubmitFNOLCommand struct {
	TenantID   string
	PolicyID   string
	LossType   string
	DateOfLoss time.Time
}

type ClaimRepository interface {
	Save(ctx context.Context, claim *domain.Claim, eventPayload []byte) error
}

type PolicyService interface {
	ValidateCoverage(ctx context.Context, policyID, lossType string, dateOfLoss time.Time) (isActive bool, isCovered bool, err error)
}

type FraudService interface {
	ScoreClaim(ctx context.Context, claim *domain.Claim) (int, error)
}

type SubmitFNOLHandler struct {
	repo   ClaimRepository
	policy PolicyService
	fraud  FraudService
}

func NewSubmitFNOLHandler(repo ClaimRepository, policy PolicyService, fraud FraudService) *SubmitFNOLHandler {
	return &SubmitFNOLHandler{
		repo:   repo,
		policy: policy,
		fraud:  fraud,
	}
}

func (h *SubmitFNOLHandler) Handle(ctx context.Context, cmd SubmitFNOLCommand) (*domain.Claim, error) {
	// 1. Instantiate Aggregate
	// Hardcoded false for vulnerability check; in reality, this is fetched via party-mgmt-svc RPC
	isVulnerable := false 
	claim := domain.NewClaim(cmd.TenantID, cmd.PolicyID, cmd.LossType, cmd.DateOfLoss, isVulnerable)

	// 2. Cross-BC Validation (Sync)
	isActive, isCovered, err := h.policy.ValidateCoverage(ctx, cmd.PolicyID, cmd.LossType, cmd.DateOfLoss)
	if err != nil {
		return nil, err // Circuit-broken
	}
	claim.ValidateCoverage(isActive, isCovered)

	// 3. Triage & Fraud
	if claim.Status == domain.Registered {
		score, _ := h.fraud.ScoreClaim(ctx, claim)
		// For simplicity, hardcode STP eligibility based on a dummy condition
		stpEligible := score < 20
		claim.Triage(score, stpEligible)
	}

	// 4. Generate Domain Event Payload (Avro mocked)
	eventPayload := []byte(`{"event_id": "mock-uuid", "claim_id": "` + claim.ID + `"}`)

	// 5. Persist Aggregate & Outbox transactionally
	err = h.repo.Save(ctx, claim, eventPayload)
	if err != nil {
		return nil, err
	}

	return claim, nil
}
