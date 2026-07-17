package aggregate

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "PENDING"
	PaymentStatusSuccess PaymentStatus = "SUCCESS"
	PaymentStatusFailed  PaymentStatus = "FAILED"
)

type Payment struct {
	ID                   uuid.UUID
	TenantID             string
	GatewayTransactionID string
	Method               string
	TotalAmount          decimal.Decimal
	UnallocatedAmount    decimal.Decimal
	Status               PaymentStatus
	Allocations          []PaymentAllocation
	CreatedAt            time.Time
}

type PaymentAllocation struct {
	ID        int64
	PaymentID uuid.UUID
	InvoiceID uuid.UUID
	Amount    decimal.Decimal
	CreatedAt time.Time
}

func NewPayment(tenantID, method, gatewayTxID string, amount decimal.Decimal) *Payment {
	return &Payment{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		GatewayTransactionID: gatewayTxID,
		Method:               method,
		TotalAmount:          amount,
		UnallocatedAmount:    amount,
		Status:               PaymentStatusPending,
		Allocations:          make([]PaymentAllocation, 0),
		CreatedAt:            time.Now().UTC(),
	}
}

func (p *Payment) MarkSuccess() {
	p.Status = PaymentStatusSuccess
}

func (p *Payment) MarkFailed() {
	p.Status = PaymentStatusFailed
}

func (p *Payment) Allocate(invoiceID uuid.UUID, amount decimal.Decimal) error {
	if p.Status != PaymentStatusSuccess {
		return errors.New("cannot allocate from a non-successful payment")
	}
	if p.UnallocatedAmount.LessThan(amount) {
		return errors.New("insufficient unallocated funds")
	}

	p.Allocations = append(p.Allocations, PaymentAllocation{
		PaymentID: p.ID,
		InvoiceID: invoiceID,
		Amount:    amount,
		CreatedAt: time.Now().UTC(),
	})
	p.UnallocatedAmount = p.UnallocatedAmount.Sub(amount)
	return nil
}
