package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-integration-svc/internal/application/ports"
	"github.com/medhen/pc-integration-svc/internal/domain"
)

type WebhookRepository struct {
	pool *pgxpool.Pool
}

func NewWebhookRepository(pool *pgxpool.Pool) ports.WebhookReceiptRepository {
	return &WebhookRepository{pool: pool}
}

func (r *WebhookRepository) SaveIfNotExists(ctx context.Context, receipt *domain.WebhookReceipt) (bool, error) {
	query := `
		INSERT INTO webhook_receipts (
			id, provider, provider_transaction_id, status, raw_payload_encrypted, received_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (provider, provider_transaction_id) DO NOTHING
	`
	tag, err := r.pool.Exec(ctx, query,
		receipt.ID, receipt.Provider, receipt.ProviderTransactionID, receipt.Status, receipt.RawPayloadEncrypted, receipt.ReceivedAt,
	)
	if err != nil {
		return false, fmt.Errorf("failed to insert webhook receipt: %w", err)
	}

	// If RowsAffected is 0, it means it already existed (duplicate).
	return tag.RowsAffected() > 0, nil
}
