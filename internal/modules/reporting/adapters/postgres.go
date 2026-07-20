// Package adapters holds the reporting module's KPI projection store.
package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	reportapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/reporting/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
)

// Schema is the DDL for the KPI projection.
const Schema = `
CREATE TABLE IF NOT EXISTS reporting_kpis (
    tenant_id             TEXT PRIMARY KEY,
    premium_written_minor BIGINT NOT NULL DEFAULT 0,
    claims_paid_minor     BIGINT NOT NULL DEFAULT 0,
    policy_count          BIGINT NOT NULL DEFAULT 0,
    claim_count           BIGINT NOT NULL DEFAULT 0
);
`

// KPIRepository implements app.Repository.
type KPIRepository struct{ db *database.DB }

// NewKPIRepository builds the repository.
func NewKPIRepository(db *database.DB) *KPIRepository { return &KPIRepository{db: db} }

var _ reportapp.Repository = (*KPIRepository)(nil)

// AddPolicy folds premium and a policy count into the tenant's projection.
func (r *KPIRepository) AddPolicy(ctx context.Context, tenantID string, grossMinor int64) error {
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO reporting_kpis (tenant_id, premium_written_minor, policy_count)
		 VALUES ($1, $2, 1)
		 ON CONFLICT (tenant_id) DO UPDATE SET
		   premium_written_minor = reporting_kpis.premium_written_minor + EXCLUDED.premium_written_minor,
		   policy_count = reporting_kpis.policy_count + 1`, tenantID, grossMinor)
	if err != nil {
		return fmt.Errorf("kpi repo: add policy: %w", err)
	}
	return nil
}

// AddClaim folds claim payment and a claim count into the tenant's projection.
func (r *KPIRepository) AddClaim(ctx context.Context, tenantID string, amountMinor int64) error {
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO reporting_kpis (tenant_id, claims_paid_minor, claim_count)
		 VALUES ($1, $2, 1)
		 ON CONFLICT (tenant_id) DO UPDATE SET
		   claims_paid_minor = reporting_kpis.claims_paid_minor + EXCLUDED.claims_paid_minor,
		   claim_count = reporting_kpis.claim_count + 1`, tenantID, amountMinor)
	if err != nil {
		return fmt.Errorf("kpi repo: add claim: %w", err)
	}
	return nil
}

// Get returns the projection for a tenant (zero-valued when none exists yet).
func (r *KPIRepository) Get(ctx context.Context, tenantID string) (reportapp.KPIRecord, error) {
	rec := reportapp.KPIRecord{TenantID: tenantID}
	err := r.db.Conn(ctx).QueryRow(ctx,
		`SELECT premium_written_minor, claims_paid_minor, policy_count, claim_count
		   FROM reporting_kpis WHERE tenant_id=$1`, tenantID).
		Scan(&rec.PremiumWrittenMinor, &rec.ClaimsPaidMinor, &rec.PolicyCount, &rec.ClaimCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return rec, nil
	}
	if err != nil {
		return rec, fmt.Errorf("kpi repo: get: %w", err)
	}
	return rec, nil
}
