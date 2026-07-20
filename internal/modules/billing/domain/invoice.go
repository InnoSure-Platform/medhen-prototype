// Package domain is the billing bounded context: invoices raised for issued
// policies and the payments applied to them. Money is platform/money throughout.
package domain

import (
	"errors"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

var (
	ErrNonPositivePayment = errors.New("billing: payment must be positive")
)

// InvoiceStatus is the invoice lifecycle state.
type InvoiceStatus string

const (
	InvoiceOpen          InvoiceStatus = "OPEN"
	InvoicePartiallyPaid InvoiceStatus = "PARTIALLY_PAID"
	InvoicePaid          InvoiceStatus = "PAID"
)

// Invoice is a demand for premium against an issued policy.
type Invoice struct {
	ID         string
	TenantID   string
	PolicyID   string
	PartyID    string
	AmountDue  money.Amount
	AmountPaid money.Amount
	Status     InvoiceStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Version    int
}

// NewInvoice raises an OPEN invoice for the full amount due.
func NewInvoice(tenantID, policyID, partyID string, due money.Amount) *Invoice {
	now := time.Now().UTC()
	return &Invoice{
		ID: ids.New(), TenantID: tenantID, PolicyID: policyID, PartyID: partyID,
		AmountDue: due, AmountPaid: money.Zero(), Status: InvoiceOpen,
		CreatedAt: now, UpdatedAt: now, Version: 1,
	}
}

// Outstanding is the amount still owed (never below zero).
func (i *Invoice) Outstanding() money.Amount {
	out := i.AmountDue.Sub(i.AmountPaid)
	if out.IsNegative() {
		return money.Zero()
	}
	return out
}

// Apply records a payment against the invoice and recomputes status. Overpayment
// is accepted (status PAID); the excess is the caller's concern (suspense).
func (i *Invoice) Apply(amount money.Amount) error {
	if amount.IsNegative() || amount.IsZero() {
		return ErrNonPositivePayment
	}
	i.AmountPaid = i.AmountPaid.Add(amount)
	switch {
	case i.AmountPaid.Cmp(i.AmountDue) >= 0:
		i.Status = InvoicePaid
	default:
		i.Status = InvoicePartiallyPaid
	}
	i.UpdatedAt = time.Now().UTC()
	i.Version++
	return nil
}

// Payment is a received payment applied to an invoice.
type Payment struct {
	ID         string
	TenantID   string
	InvoiceID  string
	Amount     money.Amount
	Method     string
	Reference  string
	ReceivedAt time.Time
}

// NewPayment builds a payment record.
func NewPayment(tenantID, invoiceID string, amount money.Amount, method, reference string) *Payment {
	return &Payment{
		ID: ids.New(), TenantID: tenantID, InvoiceID: invoiceID, Amount: amount,
		Method: method, Reference: reference, ReceivedAt: time.Now().UTC(),
	}
}
