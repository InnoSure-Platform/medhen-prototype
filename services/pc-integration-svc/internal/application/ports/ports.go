package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/medhen/pc-integration-svc/internal/domain"
)

// TransactionRepository defines the interface for saving and loading transactions.
type TransactionRepository interface {
	Save(ctx context.Context, txn *domain.IntegrationTransaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.IntegrationTransaction, error)
	UpdateState(ctx context.Context, id uuid.UUID, state domain.TransactionState) error
}

// WebhookReceiptRepository handles webhook idempotency tracking.
type WebhookReceiptRepository interface {
	SaveIfNotExists(ctx context.Context, receipt *domain.WebhookReceipt) (bool, error)
}

// EventPublisher defines how domain events are published (e.g. to Outbox).
type EventPublisher interface {
	PublishPaymentSettled(ctx context.Context, event *domain.PaymentSettledEvent) error
	PublishPaymentFailed(ctx context.Context, event *domain.PaymentFailedEvent) error
}

// PaymentProvider abstracts a third-party gateway (Telebirr, CBE, etc).
type PaymentProvider interface {
	InitiatePayment(ctx context.Context, txn *domain.IntegrationTransaction) (string, error)
}
