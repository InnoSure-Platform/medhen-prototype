// Package adapters holds the document module's Postgres repository.
package adapters

import (
	"context"
	"errors"
	"fmt"

	docapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/document/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/document/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

// Schema is the DDL for the document module's table.
const Schema = `
CREATE TABLE IF NOT EXISTS documents (
    id         TEXT PRIMARY KEY,
    tenant_id  TEXT NOT NULL,
    policy_id  TEXT NOT NULL,
    type       TEXT NOT NULL,
    number     TEXT NOT NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (tenant_id, policy_id, type)
);
`

// DocumentRepository implements app.Repository.
type DocumentRepository struct{ db *database.DB }

// NewDocumentRepository builds the repository.
func NewDocumentRepository(db *database.DB) *DocumentRepository { return &DocumentRepository{db: db} }

var _ docapp.Repository = (*DocumentRepository)(nil)

// Save inserts a document using the ambient connection.
func (r *DocumentRepository) Save(ctx context.Context, d *domain.Document) error {
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO documents (id, tenant_id, policy_id, type, number, content, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		d.ID, d.TenantID, d.PolicyID, d.Type, d.Number, d.Content, d.CreatedAt)
	if err != nil {
		return fmt.Errorf("document repo: save: %w", err)
	}
	return nil
}

func (r *DocumentRepository) scan(row pgx.Row) (*domain.Document, error) {
	var d domain.Document
	err := row.Scan(&d.ID, &d.TenantID, &d.PolicyID, &d.Type, &d.Number, &d.Content, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, docapp.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("document repo: scan: %w", err)
	}
	return &d, nil
}

const cols = `id, tenant_id, policy_id, type, number, content, created_at`

// Get loads a document by id.
func (r *DocumentRepository) Get(ctx context.Context, tenantID, id string) (*domain.Document, error) {
	return r.scan(r.db.Conn(ctx).QueryRow(ctx,
		`SELECT `+cols+` FROM documents WHERE tenant_id=$1 AND id=$2`, tenantID, id))
}

// GetByPolicy loads the certificate for a policy.
func (r *DocumentRepository) GetByPolicy(ctx context.Context, tenantID, policyID string) (*domain.Document, error) {
	return r.scan(r.db.Conn(ctx).QueryRow(ctx,
		`SELECT `+cols+` FROM documents WHERE tenant_id=$1 AND policy_id=$2 LIMIT 1`, tenantID, policyID))
}
