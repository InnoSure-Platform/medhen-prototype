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
    id             TEXT PRIMARY KEY,
    policy_number  TEXT NOT NULL UNIQUE,
    tenant_id      TEXT NOT NULL,
    quote_id       TEXT NOT NULL,
    party_id       TEXT NOT NULL,
    product_code   TEXT NOT NULL,
    gross_minor    BIGINT NOT NULL,
    status         TEXT NOT NULL,
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to   TIMESTAMPTZ NOT NULL,
    issued_at      TIMESTAMPTZ NOT NULL,
    version        INT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_policies_tenant ON policies (tenant_id);

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

// PolicyRepository implements app.PolicyRepository.
type PolicyRepository struct{ db *database.DB }

// NewPolicyRepository builds the repository.
func NewPolicyRepository(db *database.DB) *PolicyRepository { return &PolicyRepository{db: db} }

var _ policyapp.PolicyRepository = (*PolicyRepository)(nil)

// Save inserts a policy using the ambient connection.
func (r *PolicyRepository) Save(ctx context.Context, p *domain.Policy) error {
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO policies (id, policy_number, tenant_id, quote_id, party_id, product_code,
		    gross_minor, status, effective_from, effective_to, issued_at, version)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		p.ID, p.PolicyNumber, p.TenantID, p.QuoteID, p.PartyID, p.ProductCode,
		p.GrossPremium.Minor(), p.Status, p.EffectiveFrom, p.EffectiveTo, p.IssuedAt, p.Version)
	if err != nil {
		return fmt.Errorf("policy repo: save: %w", err)
	}
	return nil
}

// Get loads a policy scoped to a tenant.
func (r *PolicyRepository) Get(ctx context.Context, tenantID, id string) (*domain.Policy, error) {
	var p domain.Policy
	var gross int64
	err := r.db.Conn(ctx).QueryRow(ctx,
		`SELECT id, policy_number, tenant_id, quote_id, party_id, product_code, gross_minor,
		        status, effective_from, effective_to, issued_at, version
		   FROM policies WHERE tenant_id=$1 AND id=$2`, tenantID, id).
		Scan(&p.ID, &p.PolicyNumber, &p.TenantID, &p.QuoteID, &p.PartyID, &p.ProductCode, &gross,
			&p.Status, &p.EffectiveFrom, &p.EffectiveTo, &p.IssuedAt, &p.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, policyapp.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("policy repo: get: %w", err)
	}
	p.GrossPremium = money.FromMinor(gross)
	return &p, nil
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
