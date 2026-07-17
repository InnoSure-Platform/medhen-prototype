package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/medhen/pc-billing-svc/internal/domain/aggregate"
)

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

type BillingAccountRepository interface {
	GetByCustomerID(ctx context.Context, customerID uuid.UUID) (*aggregate.BillingAccount, error)
	GetByID(ctx context.Context, accountID uuid.UUID) (*aggregate.BillingAccount, error)
	Save(ctx context.Context, account *aggregate.BillingAccount) error
}

type InvoiceRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*aggregate.Invoice, error)
	Save(ctx context.Context, invoice *aggregate.Invoice) error
}

type PaymentRepository interface {
	GetByGatewayTxID(ctx context.Context, gatewayTxID string) (*aggregate.Payment, error)
	Save(ctx context.Context, payment *aggregate.Payment) error
}

type LedgerRepository interface {
	Save(ctx context.Context, ledger *aggregate.LedgerTransaction) error
}
