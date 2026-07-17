package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/medhen/pc-integration-svc/internal/application/ports"
	"github.com/medhen/pc-integration-svc/internal/domain"
)

type InitiatePaymentCmd struct {
	InternalReferenceID uuid.UUID
	Provider            string
	Amount              float64
	Currency            string
}

type InitiatePaymentHandler struct {
	repo      ports.TransactionRepository
	providers map[string]ports.PaymentProvider
}

func NewInitiatePaymentHandler(repo ports.TransactionRepository, providers map[string]ports.PaymentProvider) *InitiatePaymentHandler {
	return &InitiatePaymentHandler{
		repo:      repo,
		providers: providers,
	}
}

func (h *InitiatePaymentHandler) Handle(ctx context.Context, cmd InitiatePaymentCmd) (string, error) {
	provider, exists := h.providers[cmd.Provider]
	if !exists {
		return "", fmt.Errorf("provider %s is not supported", cmd.Provider)
	}

	txn := domain.NewIntegrationTransaction(cmd.InternalReferenceID, cmd.Provider, "PAYMENT", cmd.Amount, cmd.Currency)

	// Save initial state
	if err := h.repo.Save(ctx, txn); err != nil {
		return "", fmt.Errorf("failed to save transaction: %w", err)
	}

	// Call the external provider (this is where the circuit breaker is wrapped within the provider implementation)
	redirectURL, err := provider.InitiatePayment(ctx, txn)
	if err != nil {
		// Even if the circuit is open or it fails, we record the failure.
		txn.MarkFailed()
		_ = h.repo.UpdateState(ctx, txn.InternalReferenceID, txn.State)
		return "", fmt.Errorf("provider unavailable: %w", err)
	}

	// Success initiating
	txn.MarkPending()
	if err := h.repo.UpdateState(ctx, txn.InternalReferenceID, txn.State); err != nil {
		return "", fmt.Errorf("failed to update state: %w", err)
	}

	return redirectURL, nil
}
