package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/medhen/pc-billing-svc/internal/domain/aggregate"
)

type InvoiceRepository struct{}

func NewInvoiceRepository() *InvoiceRepository {
	return &InvoiceRepository{}
}

func (r *InvoiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*aggregate.Invoice, error) {
	tx := ExtractTx(ctx)
	if tx == nil {
		return nil, errors.New("no active transaction")
	}

	var inv aggregate.Invoice
	err := tx.QueryRow(ctx, `
		SELECT id, billing_account_id, policy_id, invoice_type, total_amount, amount_paid, status, due_date, coverage_start_date, coverage_end_date, created_at
		FROM invoices
		WHERE id = $1
	`, id).Scan(
		&inv.ID, &inv.BillingAccountID, &inv.PolicyID, &inv.Type,
		&inv.TotalAmount, &inv.AmountPaid, &inv.Status, &inv.DueDate, &inv.CoverageStartDate, &inv.CoverageEndDate, &inv.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	return &inv, nil
}

func (r *InvoiceRepository) Save(ctx context.Context, invoice *aggregate.Invoice) error {
	tx := ExtractTx(ctx)
	if tx == nil {
		return errors.New("no active transaction")
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO invoices (id, billing_account_id, policy_id, invoice_type, total_amount, amount_paid, status, due_date, coverage_start_date, coverage_end_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			amount_paid = EXCLUDED.amount_paid,
			status = EXCLUDED.status
	`, invoice.ID, invoice.BillingAccountID, invoice.PolicyID, invoice.Type,
		invoice.TotalAmount, invoice.AmountPaid, invoice.Status, invoice.DueDate, invoice.CoverageStartDate, invoice.CoverageEndDate, invoice.CreatedAt)

	return err
}
