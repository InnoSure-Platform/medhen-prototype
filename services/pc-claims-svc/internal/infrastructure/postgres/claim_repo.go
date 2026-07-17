package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"medhen/pc-claims-svc/internal/domain"
)

type ClaimRepository struct {
	pool *pgxpool.Pool
}

func NewClaimRepository(pool *pgxpool.Pool) *ClaimRepository {
	return &ClaimRepository{pool: pool}
}

// Save persists the Claim aggregate and writes the outbox event in the same transaction
func (r *ClaimRepository) Save(ctx context.Context, claim *domain.Claim, eventPayload []byte) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	lossDetailsJSON, err := json.Marshal(claim.LossDetails)
	if err != nil {
		return err
	}

	// 1. Upsert Claim
	_, err = tx.Exec(ctx, `
		INSERT INTO claims (id, tenant_id, claim_number, policy_id, status, version, date_of_loss, loss_type, fraud_score, loss_details, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			fraud_score = EXCLUDED.fraud_score,
			version = claims.version + 1,
			updated_at = EXCLUDED.updated_at
	`, claim.ID, claim.TenantID, claim.ClaimNumber, claim.PolicyID, claim.Status, claim.Version, 
	   claim.LossDetails.DateOfLoss, claim.LossDetails.LossType, claim.FraudScore, lossDetailsJSON, claim.CreatedAt, claim.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to save claim: %w", err)
	}

	// 2. Insert Outbox Event
	if eventPayload != nil {
		partitionKey := fmt.Sprintf("%s:%s", claim.TenantID, claim.ID)
		_, err = tx.Exec(ctx, `
			INSERT INTO outbox (topic, partition_key, payload, headers)
			VALUES ($1, $2, $3, $4)
		`, "pc.claim.lifecycle.v1", partitionKey, eventPayload, "{}")
		if err != nil {
			return fmt.Errorf("failed to save outbox event: %w", err)
		}
	}

	return tx.Commit(ctx)
}
