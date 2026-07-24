// Package app holds the billing use cases: raising an invoice when a policy is
// issued (event-driven) and applying payments. Both write their aggregate and
// their outbox event in one Unit-of-Work.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
)

// ErrNotFound is returned when an invoice does not exist.
var ErrNotFound = errors.New("billing: invoice not found")

// InvoiceRepository persists invoices (via the ambient UoW connection).
type InvoiceRepository interface {
	Save(ctx context.Context, inv *domain.Invoice) error
	Get(ctx context.Context, tenantID, id string) (*domain.Invoice, error)
	GetByPolicy(ctx context.Context, tenantID, policyID string) (*domain.Invoice, error)
}

// PaymentRepository persists payments.
type PaymentRepository interface {
	Save(ctx context.Context, p *domain.Payment) error
}

// Deps are the billing service's collaborators.
type Deps struct {
	DB       *database.DB
	Invoices InvoiceRepository
	Payments PaymentRepository
}

// Service implements the billing use cases.
type Service struct{ deps Deps }

// NewService builds the service.
func NewService(deps Deps) *Service { return &Service{deps: deps} }

// RaiseInvoiceForPolicy raises an OPEN invoice for an issued policy. It is
// idempotent: a second call for the same policy returns the existing invoice
// (so redelivery of policy.issued does not double-bill).
func (s *Service) RaiseInvoiceForPolicy(ctx context.Context, tenantID, policyID, partyID string, grossMinor int64) (*domain.Invoice, error) {
	if existing, err := s.deps.Invoices.GetByPolicy(ctx, tenantID, policyID); err == nil {
		return existing, nil
	}

	inv := domain.NewInvoice(tenantID, policyID, partyID, money.FromMinor(grossMinor))
	err := s.deps.DB.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.deps.Invoices.Save(ctx, inv); err != nil {
			return err
		}
		evt := domain.InvoiceRaised{
			InvoiceID: inv.ID, TenantID: tenantID, PolicyID: policyID,
			AmountDue: inv.AmountDue.Minor(), OccurredAt: time.Now().UTC(),
		}
		return writeEvent(ctx, s.deps.DB, domain.TopicInvoiceRaised, "invoice", inv.ID, evt)
	})
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// RecordPayment applies a payment to an invoice and emits PaymentReceived —
// invoice update, payment insert and event all in one transaction.
func (s *Service) RecordPayment(ctx context.Context, tenantID, invoiceID string, amount money.Amount, method, reference string) (*domain.Invoice, error) {
	inv, err := s.deps.Invoices.Get(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, err
	}

	err = s.deps.DB.WithinTx(ctx, func(ctx context.Context) error {
		if err := inv.Apply(amount); err != nil {
			return err
		}
		payment := domain.NewPayment(tenantID, invoiceID, amount, method, reference)
		if err := s.deps.Payments.Save(ctx, payment); err != nil {
			return err
		}
		if err := s.deps.Invoices.Save(ctx, inv); err != nil {
			return err
		}
		evt := domain.PaymentReceived{
			PaymentID: payment.ID, TenantID: tenantID, InvoiceID: invoiceID,
			Amount: amount.Minor(), OccurredAt: time.Now().UTC(),
		}
		return writeEvent(ctx, s.deps.DB, domain.TopicPaymentReceived, "invoice", invoiceID, evt)
	})
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// GetInvoice loads an invoice.
func (s *Service) GetInvoice(ctx context.Context, tenantID, id string) (*domain.Invoice, error) {
	return s.deps.Invoices.Get(ctx, tenantID, id)
}

// GetInvoiceByPolicy loads the invoice raised for a policy (so the UI can resolve
// an invoice from a policy id).
func (s *Service) GetInvoiceByPolicy(ctx context.Context, tenantID, policyID string) (*domain.Invoice, error) {
	return s.deps.Invoices.GetByPolicy(ctx, tenantID, policyID)
}

// PaymentIntent is the handle returned by InitiatePayment for a Telebirr checkout.
type PaymentIntent struct {
	InvoiceID   string `json:"invoice_id"`
	Reference   string `json:"reference"`
	AmountMinor int64  `json:"amount_minor"`
	CheckoutURL string `json:"checkout_url"`
	Status      string `json:"status"`
}

// ErrAlreadyPaid is returned when initiating payment on a settled invoice.
var ErrAlreadyPaid = errors.New("billing: invoice already paid")

// InitiatePayment starts a Telebirr checkout for an invoice's outstanding amount:
// it generates a payment reference, emits PaymentInitiated (audited), and returns
// a checkout handle. The payment is only APPLIED later by the HMAC-verified
// webhook (RecordPayment), so this is safe to call from the browser flow.
func (s *Service) InitiatePayment(ctx context.Context, tenantID, invoiceID, checkoutBase string) (*PaymentIntent, error) {
	inv, err := s.deps.Invoices.Get(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, err
	}
	if inv.Status == domain.InvoicePaid {
		return nil, ErrAlreadyPaid
	}

	outstanding := inv.Outstanding()
	reference := ids.New()
	err = s.deps.DB.WithinTx(ctx, func(ctx context.Context) error {
		return writeEvent(ctx, s.deps.DB, domain.TopicPaymentInitiated, "invoice", invoiceID, domain.PaymentInitiated{
			InvoiceID: invoiceID, TenantID: tenantID, Reference: reference,
			Amount: outstanding.Minor(), OccurredAt: time.Now().UTC(),
		})
	})
	if err != nil {
		return nil, err
	}

	if checkoutBase == "" {
		checkoutBase = "https://checkout.telebirr.et/pay"
	}
	return &PaymentIntent{
		InvoiceID:   invoiceID,
		Reference:   reference,
		AmountMinor: outstanding.Minor(),
		CheckoutURL: fmt.Sprintf("%s?ref=%s&amount=%d", checkoutBase, reference, outstanding.Minor()),
		Status:      "PENDING",
	}, nil
}

func writeEvent(ctx context.Context, db *database.DB, topic, aggType, aggID string, evt any) error {
	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("billing: marshal event: %w", err)
	}
	return outbox.Write(ctx, db.Conn(ctx), outbox.Message{
		ID: ids.New(), Topic: topic, AggregateType: aggType, AggregateID: aggID, Payload: payload,
	})
}
