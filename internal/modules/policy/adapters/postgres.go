// Package adapters holds the policy module's Postgres repositories.
package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	policyapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

// Schema is the DDL for the policy module's tables.
const Schema = `
CREATE TABLE IF NOT EXISTS quotes (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL,
    party_id        TEXT NOT NULL,
    product_code    TEXT NOT NULL,
    coverages       JSONB NOT NULL,
    risk_dimensions JSONB NOT NULL,
    net_minor       BIGINT NOT NULL,
    taxes_minor     BIGINT NOT NULL,
    gross_minor     BIGINT NOT NULL,
    calculation_id  TEXT NOT NULL,
    status          TEXT NOT NULL,
    version         INT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_quotes_tenant ON quotes (tenant_id);

CREATE TABLE IF NOT EXISTS policies (
    id              TEXT PRIMARY KEY,
    policy_number   TEXT NOT NULL UNIQUE,
    tenant_id       TEXT NOT NULL,
    quote_id        TEXT NOT NULL,
    party_id        TEXT NOT NULL,
    product_code    TEXT NOT NULL,
    gross_minor     BIGINT NOT NULL,
    status          TEXT NOT NULL,
    effective_from  TIMESTAMPTZ NOT NULL,
    effective_to    TIMESTAMPTZ NOT NULL,
    issued_at       TIMESTAMPTZ NOT NULL,
    version         INT NOT NULL,
    prior_policy_id TEXT NOT NULL DEFAULT '',
    cancel_reason   TEXT NOT NULL DEFAULT '',
    cancelled_at    TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_policies_tenant ON policies (tenant_id);
-- Servicing columns for pre-existing deployments (idempotent).
ALTER TABLE policies ADD COLUMN IF NOT EXISTS prior_policy_id TEXT NOT NULL DEFAULT '';
ALTER TABLE policies ADD COLUMN IF NOT EXISTS cancel_reason TEXT NOT NULL DEFAULT '';
ALTER TABLE policies ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS policy_sequences (
    name    TEXT PRIMARY KEY,
    current BIGINT NOT NULL
);
`

// QuoteRepository implements app.QuoteRepository.
type QuoteRepository struct{ db *database.DB }

// NewQuoteRepository builds the repository.
func NewQuoteRepository(db *database.DB) *QuoteRepository { return &QuoteRepository{db: db} }

var _ policyapp.QuoteRepository = (*QuoteRepository)(nil)

// Save upserts a quote using the ambient connection.
func (r *QuoteRepository) Save(ctx context.Context, q *domain.Quote) error {
	coverages, _ := json.Marshal(q.Coverages)
	dims, _ := json.Marshal(q.RiskDimensions)
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO quotes (id, tenant_id, party_id, product_code, coverages, risk_dimensions,
		    net_minor, taxes_minor, gross_minor, calculation_id, status, version, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		 ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, version=EXCLUDED.version,
		    updated_at=EXCLUDED.updated_at`,
		q.ID, q.TenantID, q.PartyID, q.ProductCode, coverages, dims,
		q.NetPremium.Minor(), q.TotalTaxes.Minor(), q.GrossPremium.Minor(),
		q.CalculationID, q.Status, q.Version, q.CreatedAt, q.UpdatedAt)
	if err != nil {
		return fmt.Errorf("quote repo: save: %w", err)
	}
	return nil
}

// Get loads a quote scoped to a tenant.
func (r *QuoteRepository) Get(ctx context.Context, tenantID, id string) (*domain.Quote, error) {
	var q domain.Quote
	var coverages, dims []byte
	var net, taxes, gross int64
	err := r.db.Conn(ctx).QueryRow(ctx,
		`SELECT id, tenant_id, party_id, product_code, coverages, risk_dimensions,
		        net_minor, taxes_minor, gross_minor, calculation_id, status, version, created_at, updated_at
		   FROM quotes WHERE tenant_id=$1 AND id=$2`, tenantID, id).
		Scan(&q.ID, &q.TenantID, &q.PartyID, &q.ProductCode, &coverages, &dims,
			&net, &taxes, &gross, &q.CalculationID, &q.Status, &q.Version, &q.CreatedAt, &q.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, policyapp.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("quote repo: get: %w", err)
	}
	_ = json.Unmarshal(coverages, &q.Coverages)
	_ = json.Unmarshal(dims, &q.RiskDimensions)
	q.NetPremium = money.FromMinor(net)
	q.TotalTaxes = money.FromMinor(taxes)
	q.GrossPremium = money.FromMinor(gross)
	return &q, nil
}

// List returns a tenant's quotes (newest first), paginated.
func (r *QuoteRepository) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Quote, error) {
	rows, err := r.db.Conn(ctx).Query(ctx,
		`SELECT id, tenant_id, party_id, product_code, coverages, risk_dimensions,
		        net_minor, taxes_minor, gross_minor, calculation_id, status, version, created_at, updated_at
		   FROM quotes WHERE tenant_id=$1
		  ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("quote repo: list: %w", err)
	}
	defer rows.Close()

	var out []*domain.Quote
	for rows.Next() {
		var q domain.Quote
		var coverages, dims []byte
		var net, taxes, gross int64
		if err := rows.Scan(&q.ID, &q.TenantID, &q.PartyID, &q.ProductCode, &coverages, &dims,
			&net, &taxes, &gross, &q.CalculationID, &q.Status, &q.Version, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, fmt.Errorf("quote repo: list scan: %w", err)
		}
		_ = json.Unmarshal(coverages, &q.Coverages)
		_ = json.Unmarshal(dims, &q.RiskDimensions)
		q.NetPremium = money.FromMinor(net)
		q.TotalTaxes = money.FromMinor(taxes)
		q.GrossPremium = money.FromMinor(gross)
		out = append(out, &q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("quote repo: list rows: %w", err)
	}
	return out, nil
}

// PolicyRepository implements app.PolicyRepository.
type PolicyRepository struct{ db *database.DB }

// NewPolicyRepository builds the repository.
func NewPolicyRepository(db *database.DB) *PolicyRepository { return &PolicyRepository{db: db} }

var _ policyapp.PolicyRepository = (*PolicyRepository)(nil)

// Save upserts a policy using the ambient connection. Bind inserts; servicing
// (endorse/cancel) updates the mutable columns.
func (r *PolicyRepository) Save(ctx context.Context, p *domain.Policy) error {
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO policies (id, policy_number, tenant_id, quote_id, party_id, product_code,
		    gross_minor, status, effective_from, effective_to, issued_at, version,
		    prior_policy_id, cancel_reason, cancelled_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		 ON CONFLICT (id) DO UPDATE SET
		    gross_minor=EXCLUDED.gross_minor, status=EXCLUDED.status,
		    effective_to=EXCLUDED.effective_to, version=EXCLUDED.version,
		    cancel_reason=EXCLUDED.cancel_reason, cancelled_at=EXCLUDED.cancelled_at`,
		p.ID, p.PolicyNumber, p.TenantID, p.QuoteID, p.PartyID, p.ProductCode,
		p.GrossPremium.Minor(), p.Status, p.EffectiveFrom, p.EffectiveTo, p.IssuedAt, p.Version,
		p.PriorPolicyID, p.CancelReason, p.CancelledAt)
	if err != nil {
		return fmt.Errorf("policy repo: save: %w", err)
	}
	return nil
}

const policyCols = `id, policy_number, tenant_id, quote_id, party_id, product_code, gross_minor,
	        status, effective_from, effective_to, issued_at, version,
	        prior_policy_id, cancel_reason, cancelled_at`

func scanPolicy(row pgx.Row) (*domain.Policy, error) {
	var p domain.Policy
	var gross int64
	if err := row.Scan(&p.ID, &p.PolicyNumber, &p.TenantID, &p.QuoteID, &p.PartyID, &p.ProductCode, &gross,
		&p.Status, &p.EffectiveFrom, &p.EffectiveTo, &p.IssuedAt, &p.Version,
		&p.PriorPolicyID, &p.CancelReason, &p.CancelledAt); err != nil {
		return nil, err
	}
	p.GrossPremium = money.FromMinor(gross)
	return &p, nil
}

// Get loads a policy scoped to a tenant.
func (r *PolicyRepository) Get(ctx context.Context, tenantID, id string) (*domain.Policy, error) {
	p, err := scanPolicy(r.db.Conn(ctx).QueryRow(ctx,
		`SELECT `+policyCols+` FROM policies WHERE tenant_id=$1 AND id=$2`, tenantID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, policyapp.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("policy repo: get: %w", err)
	}
	return p, nil
}

// List returns a tenant's policies (newest first), paginated.
func (r *PolicyRepository) List(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Policy, error) {
	rows, err := r.db.Conn(ctx).Query(ctx,
		`SELECT `+policyCols+` FROM policies WHERE tenant_id=$1
		  ORDER BY issued_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("policy repo: list: %w", err)
	}
	defer rows.Close()

	var out []*domain.Policy
	for rows.Next() {
		p, err := scanPolicy(rows)
		if err != nil {
			return nil, fmt.Errorf("policy repo: list scan: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("policy repo: list rows: %w", err)
	}
	return out, nil
}

// NextSequence atomically increments and returns a named policy sequence, using
// the ambient transaction so the number is reserved within the bind commit.
func (r *PolicyRepository) NextSequence(ctx context.Context, name string) (int64, error) {
	var current int64
	err := r.db.Conn(ctx).QueryRow(ctx,
		`INSERT INTO policy_sequences (name, current) VALUES ($1, 1)
		 ON CONFLICT (name) DO UPDATE SET current = policy_sequences.current + 1
		 RETURNING current`, name).Scan(&current)
	if err != nil {
		return 0, fmt.Errorf("policy repo: next sequence: %w", err)
	}
	return current, nil
}
