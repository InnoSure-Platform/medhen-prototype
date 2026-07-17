package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-audit-svc/internal/domain/audit"
)

type PostgresLedger struct {
	pool *pgxpool.Pool
}

func NewPostgresLedger(pool *pgxpool.Pool) *PostgresLedger {
	return &PostgresLedger{pool: pool}
}

func (p *PostgresLedger) AppendLeaf(ctx context.Context, entry *audit.AuditLedgerEntry) error {
	query := `
		INSERT INTO audit_merkle_leaves (
			seq_id, tenant_id, leaf_hash, digital_signature, trace_id, delta_payload
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	var payload []byte
	if entry.CryptoEnvelope != nil {
		payload = entry.CryptoEnvelope.Ciphertext
	} else {
		payload = entry.DeltaPlaintext
	}

	_, err := p.pool.Exec(ctx, query,
		entry.SeqID, entry.TenantID, entry.MerkleLeafHash, entry.DigitalSignature,
		entry.TraceID, payload,
	)

	if err != nil {
		return fmt.Errorf("failed to insert merkle leaf: %w", err)
	}

	return nil
}

func (p *PostgresLedger) UpdateMerkleRoot(ctx context.Context, epochID int64, rootHash string) error {
	query := `
		INSERT INTO audit_merkle_roots (epoch_id, root_hash)
		VALUES ($1, $2)
		ON CONFLICT (epoch_id) DO UPDATE SET root_hash = EXCLUDED.root_hash
	`
	_, err := p.pool.Exec(ctx, query, epochID, rootHash)
	return err
}

func (p *PostgresLedger) CheckLegalHold(ctx context.Context, tenantID, entityID, actorID string) (bool, error) {
	// Stub implementation
	return false, nil
}
