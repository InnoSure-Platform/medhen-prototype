// Package adapters holds the billing module's driven adapters (Postgres) and the
// Telebirr signature verifier.
package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	billingapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

// Schema is the DDL for the billing module's tables.
const Schema = `
CREATE TABLE IF NOT EXISTS invoices (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL,
    policy_id         TEXT NOT NULL,
    party_id          TEXT NOT NULL,
    amount_due_minor  BIGINT NOT NULL,
    amount_paid_minor BIGINT NOT NULL,
    status            TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL,
    updated_at        TIMESTAMPTZ NOT NULL,
    version           INT NOT NULL,
    UNIQUE (tenant_id, policy_id)
);
CREATE TABLE IF NOT EXISTS payments (
    id           TEXT PRIMARY KEY,
    tenant_id    TEXT NOT NULL,
    invoice_id   TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    method       TEXT NOT NULL,
    reference    TEXT NOT NULL,
    received_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_payments_invoice ON payments (invoice_id);
`

// InvoiceRepository implements app.InvoiceRepository.
type InvoiceRepository struct{ db *database.DB }

// NewInvoiceRepository builds the repository.
func NewInvoiceRepository(db *database.DB) *InvoiceRepository { return &InvoiceRepository{db: db} }

var _ billingapp.InvoiceRepository = (*InvoiceRepository)(nil)

// Save upserts an invoice using the ambient connection.
func (r *InvoiceRepository) Save(ctx context.Context, inv *domain.Invoice) error {
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO invoices (id, tenant_id, policy_id, party_id, amount_due_minor,
		    amount_paid_minor, status, created_at, updated_at, version)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		 ON CONFLICT (id) DO UPDATE SET amount_paid_minor=EXCLUDED.amount_paid_minor,
		    status=EXCLUDED.status, updated_at=EXCLUDED.updated_at, version=EXCLUDED.version`,
		inv.ID, inv.TenantID, inv.PolicyID, inv.PartyID, inv.AmountDue.Minor(),
		inv.AmountPaid.Minor(), inv.Status, inv.CreatedAt, inv.UpdatedAt, inv.Version)
	if err != nil {
		return fmt.Errorf("invoice repo: save: %w", err)
	}
	return nil
}

func (r *InvoiceRepository) scan(row pgx.Row) (*domain.Invoice, error) {
	var inv domain.Invoice
	var due, paid int64
	err := row.Scan(&inv.ID, &inv.TenantID, &inv.PolicyID, &inv.PartyID, &due, &paid,
		&inv.Status, &inv.CreatedAt, &inv.UpdatedAt, &inv.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, billingapp.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("invoice repo: scan: %w", err)
	}
	inv.AmountDue = money.FromMinor(due)
	inv.AmountPaid = money.FromMinor(paid)
	return &inv, nil
}

const invoiceCols = `id, tenant_id, policy_id, party_id, amount_due_minor, amount_paid_minor,
	status, created_at, updated_at, version`

// Get loads an invoice by id within a tenant.
func (r *InvoiceRepository) Get(ctx context.Context, tenantID, id string) (*domain.Invoice, error) {
	return r.scan(r.db.Conn(ctx).QueryRow(ctx,
		`SELECT `+invoiceCols+` FROM invoices WHERE tenant_id=$1 AND id=$2`, tenantID, id))
}

// GetByPolicy loads the invoice for a policy within a tenant.
func (r *InvoiceRepository) GetByPolicy(ctx context.Context, tenantID, policyID string) (*domain.Invoice, error) {
	return r.scan(r.db.Conn(ctx).QueryRow(ctx,
		`SELECT `+invoiceCols+` FROM invoices WHERE tenant_id=$1 AND policy_id=$2`, tenantID, policyID))
}

// PaymentRepository implements app.PaymentRepository.
type PaymentRepository struct{ db *database.DB }

// NewPaymentRepository builds the repository.
func NewPaymentRepository(db *database.DB) *PaymentRepository { return &PaymentRepository{db: db} }

var _ billingapp.PaymentRepository = (*PaymentRepository)(nil)

// Save inserts a payment using the ambient connection.
func (r *PaymentRepository) Save(ctx context.Context, p *domain.Payment) error {
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO payments (id, tenant_id, invoice_id, amount_minor, method, reference, received_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		p.ID, p.TenantID, p.InvoiceID, p.Amount.Minor(), p.Method, p.Reference, p.ReceivedAt)
	if err != nil {
		return fmt.Errorf("payment repo: save: %w", err)
	}
	return nil
}
