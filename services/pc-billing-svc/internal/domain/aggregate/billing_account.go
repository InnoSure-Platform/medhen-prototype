package aggregate

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AccountType string

const (
	AccountTypeDirect AccountType = "DIRECT"
	AccountTypeBroker AccountType = "BROKER"
)

type BillingAccount struct {
	ID              uuid.UUID
	TenantID        string
	CustomerID      uuid.UUID
	Type            AccountType
	SuspenseBalance decimal.Decimal
	Version         int
	CreatedAt       time.Time
}

func NewBillingAccount(tenantID string, customerID uuid.UUID, accType AccountType) *BillingAccount {
	return &BillingAccount{
		ID:              uuid.New(),
		TenantID:        tenantID,
		CustomerID:      customerID,
		Type:            accType,
		SuspenseBalance: decimal.NewFromInt(0),
		Version:         1,
		CreatedAt:       time.Now().UTC(),
	}
}

func (b *BillingAccount) AddSuspense(amount decimal.Decimal) {
	b.SuspenseBalance = b.SuspenseBalance.Add(amount)
}

func (b *BillingAccount) DeductSuspense(amount decimal.Decimal) bool {
	if b.SuspenseBalance.LessThan(amount) {
		return false
	}
	b.SuspenseBalance = b.SuspenseBalance.Sub(amount)
	return true
}
