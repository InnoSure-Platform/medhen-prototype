package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-integration-svc/internal/application/ports"
	"github.com/medhen/pc-integration-svc/internal/domain"
)

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) ports.TransactionRepository {
	return &TransactionRepository{pool: pool}
}

func (r *TransactionRepository) Save(ctx context.Context, txn *domain.IntegrationTransaction) error {
	query := `
		INSERT INTO integration_transactions (
			internal_reference_id, provider, transaction_type, amount, currency, state, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		txn.InternalReferenceID, txn.Provider, txn.TransactionType, txn.Amount, txn.Currency, txn.State, txn.CreatedAt, txn.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.IntegrationTransaction, error) {
	query := `
		SELECT internal_reference_id, provider, transaction_type, amount, currency, state, created_at, updated_at
		FROM integration_transactions
		WHERE internal_reference_id = $1
	`
	var txn domain.IntegrationTransaction
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&txn.InternalReferenceID, &txn.Provider, &txn.TransactionType, &txn.Amount, &txn.Currency, &txn.State, &txn.CreatedAt, &txn.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}
	return &txn, nil
}

func (r *TransactionRepository) UpdateState(ctx context.Context, id uuid.UUID, state domain.TransactionState) error {
	query := `UPDATE integration_transactions SET state = $1, updated_at = now() WHERE internal_reference_id = $2`
	_, err := r.pool.Exec(ctx, query, state, id)
	if err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}
	return nil
}
