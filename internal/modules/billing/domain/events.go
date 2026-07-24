package domain

import "time"

// Topics published by the billing module.
const (
	TopicInvoiceRaised    = "billing.invoice_raised"
	TopicPaymentReceived  = "billing.payment_received"
	TopicPaymentInitiated = "billing.payment_initiated"
)

// PaymentInitiated is emitted when a Telebirr checkout is started for an invoice.
// Settlement still happens via the (HMAC-verified) webhook applying the payment.
type PaymentInitiated struct {
	InvoiceID  string    `json:"invoice_id"`
	TenantID   string    `json:"tenant_id"`
	Reference  string    `json:"reference"`
	Amount     int64     `json:"amount_minor"`
	OccurredAt time.Time `json:"occurred_at"`
}

// EventName satisfies eventbus.Event.
func (PaymentInitiated) EventName() string { return TopicPaymentInitiated }

// InvoiceRaised is emitted when an invoice is raised for an issued policy.
type InvoiceRaised struct {
	InvoiceID  string    `json:"invoice_id"`
	TenantID   string    `json:"tenant_id"`
	PolicyID   string    `json:"policy_id"`
	AmountDue  int64     `json:"amount_due_minor"`
	OccurredAt time.Time `json:"occurred_at"`
}

// EventName satisfies eventbus.Event.
func (InvoiceRaised) EventName() string { return TopicInvoiceRaised }

// PaymentReceived is emitted when a payment is applied to an invoice.
type PaymentReceived struct {
	PaymentID  string    `json:"payment_id"`
	TenantID   string    `json:"tenant_id"`
	InvoiceID  string    `json:"invoice_id"`
	Amount     int64     `json:"amount_minor"`
	OccurredAt time.Time `json:"occurred_at"`
}

// EventName satisfies eventbus.Event.
func (PaymentReceived) EventName() string { return TopicPaymentReceived }
