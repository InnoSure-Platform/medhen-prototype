package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-policy-svc/internal/domain/policy"
	"github.com/medhen/pc-policy-svc/internal/domain/quote"
)

type BindPolicyCommand struct {
	QuoteID uuid.UUID
}

type BindPolicyHandler struct {
	quoteRepo  quote.Repository
	policyRepo policy.Repository
	// In a real app we'd inject Outbox publisher here for events, and a Billing client for the saga
}

func NewBindPolicyHandler(quoteRepo quote.Repository, policyRepo policy.Repository) *BindPolicyHandler {
	return &BindPolicyHandler{
		quoteRepo:  quoteRepo,
		policyRepo: policyRepo,
	}
}

func (h *BindPolicyHandler) Handle(ctx context.Context, cmd BindPolicyCommand) (*policy.Policy, error) {
	q, err := h.quoteRepo.GetByID(ctx, cmd.QuoteID)
	if err != nil {
		return nil, err
	}

	if q.Status != quote.StatusAccepted {
		return nil, quote.ErrQuoteNotAccepted
	}

	// 1. Trigger Billing Saga (Mocked here)
	// If billing fails, we would return an error

	// 2. Mark Quote as Bound
	err = q.MarkBound()
	if err != nil {
		return nil, err
	}

	// 3. Create Policy
	policyNumber := fmt.Sprintf("EIC/MOT/2026/%s", q.ID.String()[:8]) // Simplistic generator
	effectiveFrom := time.Now()
	effectiveTo := effectiveFrom.AddDate(1, 0, 0)

	p, err := policy.NewPolicy(q.TenantID, policyNumber, q.ProductID, q.PartyID, q.RiskPayload, q.Premium, effectiveFrom, effectiveTo)
	if err != nil {
		return nil, err
	}

	// Save Quote and Policy transactionally (Here we just save sequentially for brevity)
	err = h.quoteRepo.Save(ctx, q)
	if err != nil {
		return nil, err
	}

	err = h.policyRepo.Save(ctx, p)
	if err != nil {
		return nil, err
	}

	// Publish Event via Outbox...

	return p, nil
}
