package aggregate

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Commission tracks the payout or net-deduction for a Broker.
type Commission struct {
	ID               uuid.UUID
	BillingAccountID uuid.UUID // Must be a BROKER account
	InvoiceID        uuid.UUID
	GrossPremium     decimal.Decimal
	CommissionRate   decimal.Decimal // e.g., 0.15 for 15%
	CommissionAmount decimal.Decimal
	NetPremium       decimal.Decimal
	CreatedAt        time.Time
}

func NewCommission(brokerAccountID, invoiceID uuid.UUID, grossPremium, rate decimal.Decimal) (*Commission, error) {
	if rate.LessThan(decimal.Zero) || rate.GreaterThan(decimal.NewFromInt(1)) {
		return nil, errors.New("commission rate must be between 0 and 1")
	}

	commissionAmount := grossPremium.Mul(rate)
	netPremium := grossPremium.Sub(commissionAmount)

	return &Commission{
		ID:               uuid.New(),
		BillingAccountID: brokerAccountID,
		InvoiceID:        invoiceID,
		GrossPremium:     grossPremium,
		CommissionRate:   rate,
		CommissionAmount: commissionAmount,
		NetPremium:       netPremium,
		CreatedAt:        time.Now().UTC(),
	}, nil
}
