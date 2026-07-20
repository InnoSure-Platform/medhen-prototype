// Package adapters holds the claims module's Postgres repository.
package adapters

import (
	"context"
	"errors"
	"fmt"

	claimsapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/jackc/pgx/v5"
)

// Schema is the DDL for the claims module's tables.
const Schema = `
CREATE TABLE IF NOT EXISTS claims (
    id             TEXT PRIMARY KEY,
    tenant_id      TEXT NOT NULL,
    policy_id      TEXT NOT NULL,
    party_id       TEXT NOT NULL,
    status         TEXT NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    latitude       DOUBLE PRECISION NOT NULL DEFAULT 0,
    longitude      DOUBLE PRECISION NOT NULL DEFAULT 0,
    reserve_minor  BIGINT NOT NULL,
    settled_minor  BIGINT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL,
    updated_at     TIMESTAMPTZ NOT NULL,
    version        INT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_claims_tenant ON claims (tenant_id);
CREATE INDEX IF NOT EXISTS idx_claims_policy ON claims (policy_id);
`

// ClaimRepository implements app.ClaimRepository.
type ClaimRepository struct{ db *database.DB }

// NewClaimRepository builds the repository.
func NewClaimRepository(db *database.DB) *ClaimRepository { return &ClaimRepository{db: db} }

var _ claimsapp.ClaimRepository = (*ClaimRepository)(nil)

// Save upserts a claim using the ambient connection.
func (r *ClaimRepository) Save(ctx context.Context, c *domain.Claim) error {
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO claims (id, tenant_id, policy_id, party_id, status, description,
		    latitude, longitude, reserve_minor, settled_minor, created_at, updated_at, version)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		 ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, settled_minor=EXCLUDED.settled_minor,
		    reserve_minor=EXCLUDED.reserve_minor, updated_at=EXCLUDED.updated_at, version=EXCLUDED.version`,
		c.ID, c.TenantID, c.PolicyID, c.PartyID, c.Status, c.Description,
		c.Latitude, c.Longitude, c.Reserve.Minor(), c.SettledAmount.Minor(),
		c.CreatedAt, c.UpdatedAt, c.Version)
	if err != nil {
		return fmt.Errorf("claim repo: save: %w", err)
	}
	return nil
}

// Get loads a claim scoped to a tenant.
func (r *ClaimRepository) Get(ctx context.Context, tenantID, id string) (*domain.Claim, error) {
	var c domain.Claim
	var reserve, settled int64
	err := r.db.Conn(ctx).QueryRow(ctx,
		`SELECT id, tenant_id, policy_id, party_id, status, description, latitude, longitude,
		        reserve_minor, settled_minor, created_at, updated_at, version
		   FROM claims WHERE tenant_id=$1 AND id=$2`, tenantID, id).
		Scan(&c.ID, &c.TenantID, &c.PolicyID, &c.PartyID, &c.Status, &c.Description,
			&c.Latitude, &c.Longitude, &reserve, &settled, &c.CreatedAt, &c.UpdatedAt, &c.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, claimsapp.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("claim repo: get: %w", err)
	}
	c.Reserve = money.FromMinor(reserve)
	c.SettledAmount = money.FromMinor(settled)
	return &c, nil
}
