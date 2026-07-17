package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/medhen/pc-billing-svc/internal/domain/aggregate"
)

type PaymentRepository struct{}

func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{}
}

func (r *PaymentRepository) GetByGatewayTxID(ctx context.Context, gatewayTxID string) (*aggregate.Payment, error) {
	tx := ExtractTx(ctx)
	if tx == nil {
		return nil, errors.New("no active transaction")
	}

	var payment aggregate.Payment
	err := tx.QueryRow(ctx, `
		SELECT id, tenant_id, gateway_transaction_id, method, total_amount, unallocated_amount, status, created_at
		FROM payments
		WHERE gateway_transaction_id = $1
	`, gatewayTxID).Scan(
		&payment.ID, &payment.TenantID, &payment.GatewayTransactionID, &payment.Method,
		&payment.TotalAmount, &payment.UnallocatedAmount, &payment.Status, &payment.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil on not found
		}
		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) Save(ctx context.Context, payment *aggregate.Payment) error {
	tx := ExtractTx(ctx)
	if tx == nil {
		return errors.New("no active transaction")
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO payments (id, tenant_id, gateway_transaction_id, method, total_amount, unallocated_amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			unallocated_amount = EXCLUDED.unallocated_amount,
			status = EXCLUDED.status
	`, payment.ID, payment.TenantID, payment.GatewayTransactionID, payment.Method,
		payment.TotalAmount, payment.UnallocatedAmount, payment.Status, payment.CreatedAt)

	return err
}
