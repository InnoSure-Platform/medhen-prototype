// Package adapters contains the party module's driven adapters (Postgres).
package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
)

// Schema is the DDL for the party module's tables (its own schema-per-module).
const Schema = `
CREATE TABLE IF NOT EXISTS parties (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL,
    type              TEXT NOT NULL,
    status            TEXT NOT NULL,
    full_name         TEXT NOT NULL,
    full_name_amharic TEXT NOT NULL DEFAULT '',
    phone_e164        TEXT NOT NULL,
    national_id       TEXT NOT NULL,
    region            TEXT NOT NULL,
    zone              TEXT NOT NULL,
    woreda            TEXT NOT NULL,
    kebele            TEXT NOT NULL DEFAULT '',
    house_number      TEXT NOT NULL DEFAULT '',
    version           INT  NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL,
    updated_at        TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_parties_tenant ON parties (tenant_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_parties_tenant_nid ON parties (tenant_id, national_id);
`

// PartyRepository implements app.Repository over Postgres.
type PartyRepository struct {
	db *database.DB
}

// NewPartyRepository builds the repository.
func NewPartyRepository(db *database.DB) *PartyRepository {
	return &PartyRepository{db: db}
}

var _ app.Repository = (*PartyRepository)(nil)

// Save upserts a party using the ambient connection (the caller's transaction).
func (r *PartyRepository) Save(ctx context.Context, p *domain.Party) error {
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO parties (id, tenant_id, type, status, full_name, full_name_amharic,
		    phone_e164, national_id, region, zone, woreda, kebele, house_number,
		    version, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		 ON CONFLICT (id) DO UPDATE SET
		    status=EXCLUDED.status, full_name=EXCLUDED.full_name,
		    full_name_amharic=EXCLUDED.full_name_amharic, phone_e164=EXCLUDED.phone_e164,
		    region=EXCLUDED.region, zone=EXCLUDED.zone, woreda=EXCLUDED.woreda,
		    kebele=EXCLUDED.kebele, house_number=EXCLUDED.house_number,
		    version=EXCLUDED.version, updated_at=EXCLUDED.updated_at`,
		p.ID, p.TenantID, p.Type, p.Status, p.FullName, p.FullNameAmharic,
		p.PhoneE164, p.NationalID, p.Address.Region, p.Address.Zone, p.Address.Woreda,
		p.Address.Kebele, p.Address.HouseNumber, p.Version, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("party repo: save: %w", err)
	}
	return nil
}

// GetByID loads a party scoped to a tenant.
func (r *PartyRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Party, error) {
	var p domain.Party
	err := r.db.Conn(ctx).QueryRow(ctx,
		`SELECT id, tenant_id, type, status, full_name, full_name_amharic, phone_e164,
		        national_id, region, zone, woreda, kebele, house_number, version,
		        created_at, updated_at
		   FROM parties WHERE tenant_id=$1 AND id=$2`, tenantID, id).
		Scan(&p.ID, &p.TenantID, &p.Type, &p.Status, &p.FullName, &p.FullNameAmharic,
			&p.PhoneE164, &p.NationalID, &p.Address.Region, &p.Address.Zone, &p.Address.Woreda,
			&p.Address.Kebele, &p.Address.HouseNumber, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, app.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("party repo: get: %w", err)
	}
	return &p, nil
}
