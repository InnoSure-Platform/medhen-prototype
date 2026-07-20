// Package adapters holds the audit module's append-only Postgres store.
package adapters

import (
	"context"
	"fmt"

	auditapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/audit/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
)

// Schema is the DDL for the audit trail. The table is append-only by convention
// (no UPDATE/DELETE paths exist in the repository).
const Schema = `
CREATE TABLE IF NOT EXISTS audit_log (
    id          TEXT PRIMARY KEY,
    topic       TEXT NOT NULL,
    tenant_id   TEXT NOT NULL DEFAULT '',
    payload     JSONB NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_audit_tenant_time ON audit_log (tenant_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_topic ON audit_log (topic);
`

// AuditRepository implements app.Repository over Postgres.
type AuditRepository struct{ db *database.DB }

// NewAuditRepository builds the repository.
func NewAuditRepository(db *database.DB) *AuditRepository { return &AuditRepository{db: db} }

var _ auditapp.Repository = (*AuditRepository)(nil)

// Append inserts an immutable trail entry using the ambient connection (so it
// commits within the outbox-relay transaction that published the event).
func (r *AuditRepository) Append(ctx context.Context, rec auditapp.Record) error {
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO audit_log (id, topic, tenant_id, payload, recorded_at)
		 VALUES ($1,$2,$3,$4,$5)`,
		rec.ID, rec.Topic, rec.TenantID, []byte(rec.Payload), rec.RecordedAt)
	if err != nil {
		return fmt.Errorf("audit repo: append: %w", err)
	}
	return nil
}

// List returns recent entries, optionally filtered by topic, newest first.
func (r *AuditRepository) List(ctx context.Context, tenantID, topic string, limit int) ([]auditapp.Record, error) {
	rows, err := r.db.Conn(ctx).Query(ctx,
		`SELECT id, topic, tenant_id, payload, recorded_at
		   FROM audit_log
		  WHERE tenant_id = $1 AND ($2 = '' OR topic = $2)
		  ORDER BY recorded_at DESC
		  LIMIT $3`, tenantID, topic, limit)
	if err != nil {
		return nil, fmt.Errorf("audit repo: list: %w", err)
	}
	defer rows.Close()

	var out []auditapp.Record
	for rows.Next() {
		var rec auditapp.Record
		var payload []byte
		if err := rows.Scan(&rec.ID, &rec.Topic, &rec.TenantID, &payload, &rec.RecordedAt); err != nil {
			return nil, err
		}
		rec.Payload = payload
		out = append(out, rec)
	}
	return out, rows.Err()
}
