package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-integration-svc/internal/application/ports"
	"github.com/medhen/pc-integration-svc/internal/domain"
)

type IngestWebhookCmd struct {
	Provider              string
	ProviderTransactionID string
	RawPayload            []byte
	StatusIsSuccess       bool
	InternalReferenceID   uuid.UUID
	Amount                float64
	Currency              string
}

type IngestWebhookHandler struct {
	webhookRepo   ports.WebhookReceiptRepository
	txnRepo       ports.TransactionRepository
	eventPub      ports.EventPublisher
}

func NewIngestWebhookHandler(
	webhookRepo ports.WebhookReceiptRepository,
	txnRepo ports.TransactionRepository,
	eventPub ports.EventPublisher,
) *IngestWebhookHandler {
	return &IngestWebhookHandler{
		webhookRepo:   webhookRepo,
		txnRepo:       txnRepo,
		eventPub:      eventPub,
	}
}

func (h *IngestWebhookHandler) Handle(ctx context.Context, cmd IngestWebhookCmd) error {
	// 1. Enforce Idempotency
	receipt := domain.NewWebhookReceipt(cmd.Provider, cmd.ProviderTransactionID, cmd.RawPayload, domain.WebhookStatusProcessed)
	inserted, err := h.webhookRepo.SaveIfNotExists(ctx, receipt)
	if err != nil {
		return fmt.Errorf("failed to check idempotency: %w", err)
	}
	if !inserted {
		// Idempotent success (200 OK)
		return nil
	}

	// 2. Load Transaction
	txn, err := h.txnRepo.GetByID(ctx, cmd.InternalReferenceID)
	if err != nil {
		return fmt.Errorf("failed to load transaction %s: %w", cmd.InternalReferenceID, err)
	}

	// 3. State Transition and Event Publishing
	if cmd.StatusIsSuccess {
		if err := txn.MarkSuccess(); err != nil {
			return fmt.Errorf("state transition error: %w", err)
		}
		
		event := &domain.PaymentSettledEvent{
			EventID:               uuid.New(),
			InternalReferenceID:   txn.InternalReferenceID,
			Provider:              txn.Provider,
			ProviderTransactionID: cmd.ProviderTransactionID,
			AmountSettled:         cmd.Amount,
			Currency:              cmd.Currency,
			SettledAt:             time.Now(),
		}
		if err := h.eventPub.PublishPaymentSettled(ctx, event); err != nil {
			return fmt.Errorf("failed to publish settled event: %w", err)
		}
	} else {
		if err := txn.MarkFailed(); err != nil {
			return fmt.Errorf("state transition error: %w", err)
		}

		event := &domain.PaymentFailedEvent{
			EventID:             uuid.New(),
			InternalReferenceID: txn.InternalReferenceID,
			Provider:            txn.Provider,
			Reason:              "Provider callback indicated failure",
			FailedAt:            time.Now(),
		}
		if err := h.eventPub.PublishPaymentFailed(ctx, event); err != nil {
			return fmt.Errorf("failed to publish failed event: %w", err)
		}
	}

	// 4. Persist updated transaction state
	if err := h.txnRepo.UpdateState(ctx, txn.InternalReferenceID, txn.State); err != nil {
		return fmt.Errorf("failed to update transaction state: %w", err)
	}

	return nil
}
