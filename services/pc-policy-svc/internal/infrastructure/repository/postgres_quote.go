package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-policy-svc/internal/domain/quote"
)

type PostgresQuoteRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresQuoteRepository(pool *pgxpool.Pool) *PostgresQuoteRepository {
	return &PostgresQuoteRepository{pool: pool}
}

func (r *PostgresQuoteRepository) Save(ctx context.Context, q *quote.Quote) error {
	payload, err := json.Marshal(q.RiskPayload)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO quotes (id, tenant_id, product_id, party_id, status, risk_payload, premium, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			premium = EXCLUDED.premium,
			expires_at = EXCLUDED.expires_at,
			updated_at = EXCLUDED.updated_at
	`
	
	_, err = r.pool.Exec(ctx, query,
		q.ID, q.TenantID, q.ProductID, q.PartyID, string(q.Status), payload, q.Premium, q.ExpiresAt, q.CreatedAt, q.UpdatedAt,
	)
	return err
}

func (r *PostgresQuoteRepository) GetByID(ctx context.Context, id uuid.UUID) (*quote.Quote, error) {
	query := `SELECT id, tenant_id, product_id, party_id, status, risk_payload, premium, expires_at, created_at, updated_at FROM quotes WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	var q quote.Quote
	var status string
	var payload []byte

	err := row.Scan(&q.ID, &q.TenantID, &q.ProductID, &q.PartyID, &status, &payload, &q.Premium, &q.ExpiresAt, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("quote not found")
		}
		return nil, err
	}

	q.Status = quote.Status(status)
	q.RiskPayload = payload

	return &q, nil
}
