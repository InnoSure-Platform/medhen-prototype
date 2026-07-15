package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository persists aggregates in PostgreSQL (per-schema isolation).
type PostgresRepository struct {
	pool   *pgxpool.Pool
	schema string
}

func NewPostgres(ctx context.Context, dsn, schema string) (*PostgresRepository, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if schema == "" {
		schema = "public"
	}
	r := &PostgresRepository{pool: pool, schema: schema}
	if err := r.EnsureSchema(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return r, nil
}

func (r *PostgresRepository) Close() { r.pool.Close() }

func (r *PostgresRepository) EnsureSchema(ctx context.Context) error {
	ddl := fmt.Sprintf(`
CREATE SCHEMA IF NOT EXISTS %s;
SET search_path TO %s;

CREATE TABLE IF NOT EXISTS parties (
  id UUID PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  full_name TEXT NOT NULL,
  full_name_am TEXT,
  phone_e164 TEXT NOT NULL,
  email TEXT,
  status TEXT NOT NULL,
  address JSONB,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS quotes (
  id UUID PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  party_id UUID NOT NULL,
  product_code TEXT NOT NULL,
  status TEXT NOT NULL,
  risk JSONB NOT NULL,
  lines JSONB NOT NULL,
  total_minor BIGINT NOT NULL,
  currency TEXT NOT NULL,
  uw_decision TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS policies (
  id UUID PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  policy_number TEXT NOT NULL,
  quote_id UUID NOT NULL,
  party_id UUID NOT NULL,
  product_code TEXT NOT NULL,
  status TEXT NOT NULL,
  risk JSONB NOT NULL,
  lines JSONB NOT NULL,
  total_minor BIGINT NOT NULL,
  currency TEXT NOT NULL,
  effective_from DATE NOT NULL,
  effective_to DATE NOT NULL,
  issued_at TIMESTAMPTZ,
  invoice_id UUID
);

CREATE TABLE IF NOT EXISTS policy_seq (
  year INT PRIMARY KEY,
  seq INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS invoices (
  id UUID PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  policy_id UUID NOT NULL,
  amount_minor BIGINT NOT NULL,
  currency TEXT NOT NULL,
  status TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS receipts (
  id TEXT PRIMARY KEY,
  invoice_id UUID NOT NULL,
  channel TEXT NOT NULL,
  status TEXT NOT NULL,
  paid_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS documents (
  id UUID PRIMARY KEY,
  policy_id UUID NOT NULL,
  doc_type TEXT NOT NULL,
  locale TEXT NOT NULL,
  url TEXT NOT NULL,
  object_key TEXT
);

CREATE TABLE IF NOT EXISTS claims (
  id UUID PRIMARY KEY,
  claim_number TEXT NOT NULL,
  tenant_id TEXT NOT NULL,
  policy_id UUID NOT NULL,
  status TEXT NOT NULL,
  track TEXT NOT NULL,
  description TEXT NOT NULL,
  latitude DOUBLE PRECISION,
  longitude DOUBLE PRECISION,
  estimated_amount_minor BIGINT,
  settlement_minor BIGINT,
  currency TEXT NOT NULL,
  photo_keys JSONB,
  created_at TIMESTAMPTZ NOT NULL,
  settled_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS claim_seq (
  year INT PRIMARY KEY,
  seq INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS audit_log (
  id UUID PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  action TEXT NOT NULL,
  actor TEXT NOT NULL,
  detail TEXT,
  at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS outbox (
  id UUID PRIMARY KEY,
  aggregate_type TEXT NOT NULL,
  aggregate_id TEXT NOT NULL,
  event_type TEXT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS outbox_unpublished_idx ON outbox (created_at) WHERE published_at IS NULL;
`, r.schema, r.schema)
	_, err := r.pool.Exec(ctx, ddl)
	return err
}

func (r *PostgresRepository) SaveParty(ctx context.Context, p *Party) error {
	addr, _ := json.Marshal(p.Address)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO parties (id, tenant_id, full_name, full_name_am, phone_e164, email, status, address, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET full_name=EXCLUDED.full_name, phone_e164=EXCLUDED.phone_e164`,
		p.ID, p.TenantID, p.FullName, p.FullNameAm, p.PhoneE164, p.Email, p.Status, addr, p.CreatedAt)
	return err
}

func (r *PostgresRepository) GetParty(ctx context.Context, id string) (*Party, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, full_name, full_name_am, phone_e164, email, status, address, created_at FROM parties WHERE id=$1`, id)
	var p Party
	var addr []byte
	if err := row.Scan(&p.ID, &p.TenantID, &p.FullName, &p.FullNameAm, &p.PhoneE164, &p.Email, &p.Status, &addr, &p.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if len(addr) > 0 {
		_ = json.Unmarshal(addr, &p.Address)
	}
	return &p, nil
}

func (r *PostgresRepository) SaveQuote(ctx context.Context, q *Quote) error {
	risk, _ := json.Marshal(q.Risk)
	lines, _ := json.Marshal(q.Lines)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO quotes (id, tenant_id, party_id, product_code, status, risk, lines, total_minor, currency, uw_decision, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status`,
		q.ID, q.TenantID, q.PartyID, q.ProductCode, q.Status, risk, lines, q.TotalMinor, q.Currency, q.UWDecision, q.ExpiresAt, q.CreatedAt)
	return err
}

func (r *PostgresRepository) GetQuote(ctx context.Context, id string) (*Quote, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, party_id, product_code, status, risk, lines, total_minor, currency, uw_decision, expires_at, created_at FROM quotes WHERE id=$1`, id)
	var q Quote
	var risk, lines []byte
	if err := row.Scan(&q.ID, &q.TenantID, &q.PartyID, &q.ProductCode, &q.Status, &risk, &lines, &q.TotalMinor, &q.Currency, &q.UWDecision, &q.ExpiresAt, &q.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal(risk, &q.Risk)
	_ = json.Unmarshal(lines, &q.Lines)
	return &q, nil
}

func (r *PostgresRepository) UpdateQuoteStatus(ctx context.Context, id, status string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE quotes SET status=$2 WHERE id=$1`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) SavePolicy(ctx context.Context, p *Policy) error {
	risk, _ := json.Marshal(p.Risk)
	lines, _ := json.Marshal(p.Lines)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO policies (id, tenant_id, policy_number, quote_id, party_id, product_code, status, risk, lines, total_minor, currency, effective_from, effective_to, issued_at, invoice_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, policy_number=EXCLUDED.policy_number, issued_at=EXCLUDED.issued_at`,
		p.ID, p.TenantID, p.PolicyNumber, p.QuoteID, p.PartyID, p.ProductCode, p.Status, risk, lines, p.TotalMinor, p.Currency, p.EffectiveFrom, p.EffectiveTo, p.IssuedAt, p.InvoiceID)
	return err
}

func (r *PostgresRepository) GetPolicy(ctx context.Context, id string) (*Policy, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, policy_number, quote_id, party_id, product_code, status, risk, lines, total_minor, currency, effective_from, effective_to, issued_at, invoice_id FROM policies WHERE id=$1`, id)
	var p Policy
	var risk, lines []byte
	var issuedAt *time.Time
	if err := row.Scan(&p.ID, &p.TenantID, &p.PolicyNumber, &p.QuoteID, &p.PartyID, &p.ProductCode, &p.Status, &risk, &lines, &p.TotalMinor, &p.Currency, &p.EffectiveFrom, &p.EffectiveTo, &issuedAt, &p.InvoiceID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal(risk, &p.Risk)
	_ = json.Unmarshal(lines, &p.Lines)
	p.IssuedAt = issuedAt
	return &p, nil
}

func (r *PostgresRepository) NextPolicyNumber(ctx context.Context, year int) (string, error) {
	var seq int
	err := r.pool.QueryRow(ctx, `
		INSERT INTO policy_seq (year, seq) VALUES ($1, 1)
		ON CONFLICT (year) DO UPDATE SET seq = policy_seq.seq + 1
		RETURNING seq`, year).Scan(&seq)
	if err != nil {
		return "", err
	}
	return formatSeq("EIC/MOT", year, seq), nil
}

func (r *PostgresRepository) ListIssuedPolicies(ctx context.Context) ([]Policy, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, tenant_id, policy_number, quote_id, party_id, product_code, status, risk, lines, total_minor, currency, effective_from, effective_to, issued_at, invoice_id FROM policies WHERE status='ISSUED'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPolicies(rows)
}

func (r *PostgresRepository) SaveInvoice(ctx context.Context, inv *Invoice) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO invoices (id, tenant_id, policy_id, amount_minor, currency, status)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status`,
		inv.ID, inv.TenantID, inv.PolicyID, inv.AmountMinor, inv.Currency, inv.Status)
	return err
}

func (r *PostgresRepository) GetInvoice(ctx context.Context, id string) (*Invoice, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, policy_id, amount_minor, currency, status FROM invoices WHERE id=$1`, id)
	var inv Invoice
	if err := row.Scan(&inv.ID, &inv.TenantID, &inv.PolicyID, &inv.AmountMinor, &inv.Currency, &inv.Status); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &inv, nil
}

func (r *PostgresRepository) SaveReceipt(ctx context.Context, rec *Receipt) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO receipts (id, invoice_id, channel, status, paid_at) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (id) DO NOTHING`,
		rec.ID, rec.InvoiceID, rec.Channel, rec.Status, rec.PaidAt)
	return err
}

func (r *PostgresRepository) SaveDocument(ctx context.Context, d *Document) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO documents (id, policy_id, doc_type, locale, url, object_key) VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (id) DO NOTHING`,
		d.ID, d.PolicyID, d.Type, d.Locale, d.URL, d.ObjectKey)
	return err
}

func (r *PostgresRepository) ListDocumentsByPolicy(ctx context.Context, policyID string) ([]Document, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, policy_id, doc_type, locale, url, object_key FROM documents WHERE policy_id=$1`, policyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Document
	for rows.Next() {
		var d Document
		var key *string
		if err := rows.Scan(&d.ID, &d.PolicyID, &d.Type, &d.Locale, &d.URL, &key); err != nil {
			return nil, err
		}
		if key != nil {
			d.ObjectKey = *key
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) SaveClaim(ctx context.Context, c *Claim) error {
	photos, _ := json.Marshal(c.PhotoObjectKeys)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO claims (id, claim_number, tenant_id, policy_id, status, track, description, latitude, longitude, estimated_amount_minor, settlement_minor, currency, photo_keys, created_at, settled_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, settlement_minor=EXCLUDED.settlement_minor, settled_at=EXCLUDED.settled_at`,
		c.ID, c.ClaimNumber, c.TenantID, c.PolicyID, c.Status, c.Track, c.Description, c.Latitude, c.Longitude, c.EstimatedAmountMinor, c.SettlementMinor, c.Currency, photos, c.CreatedAt, c.SettledAt)
	return err
}

func (r *PostgresRepository) GetClaim(ctx context.Context, id string) (*Claim, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, claim_number, tenant_id, policy_id, status, track, description, latitude, longitude, estimated_amount_minor, settlement_minor, currency, photo_keys, created_at, settled_at FROM claims WHERE id=$1`, id)
	return scanClaimRow(row)
}

func (r *PostgresRepository) NextClaimNumber(ctx context.Context, year int) (string, error) {
	var seq int
	err := r.pool.QueryRow(ctx, `
		INSERT INTO claim_seq (year, seq) VALUES ($1, 1)
		ON CONFLICT (year) DO UPDATE SET seq = claim_seq.seq + 1
		RETURNING seq`, year).Scan(&seq)
	if err != nil {
		return "", err
	}
	return formatSeq("EIC/CLM", year, seq), nil
}

func (r *PostgresRepository) ListClaims(ctx context.Context) ([]Claim, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, claim_number, tenant_id, policy_id, status, track, description, latitude, longitude, estimated_amount_minor, settlement_minor, currency, photo_keys, created_at, settled_at FROM claims`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Claim
	for rows.Next() {
		c, err := scanClaimRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) AppendAudit(ctx context.Context, e AuditEntry) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO audit_log (id, tenant_id, entity_type, entity_id, action, actor, detail, at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		e.ID, e.TenantID, e.EntityType, e.EntityID, e.Action, e.Actor, e.Detail, e.At)
	return err
}

func (r *PostgresRepository) QueryAudit(ctx context.Context, entityType, entityID string, limit int) ([]AuditEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT id, tenant_id, entity_type, entity_id, action, actor, detail, at FROM audit_log WHERE 1=1`
	args := []any{}
	i := 1
	if entityType != "" {
		q += fmt.Sprintf(" AND entity_type=$%d", i)
		args = append(args, entityType)
		i++
	}
	if entityID != "" {
		q += fmt.Sprintf(" AND entity_id=$%d", i)
		args = append(args, entityID)
		i++
	}
	q += fmt.Sprintf(" ORDER BY at DESC LIMIT $%d", i)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.TenantID, &e.EntityType, &e.EntityID, &e.Action, &e.Actor, &e.Detail, &e.At); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) InsertOutbox(ctx context.Context, aggregateType, aggregateID, eventType string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO outbox (id, aggregate_type, aggregate_id, event_type, payload) VALUES ($1,$2,$3,$4,$5)`,
		uuid.NewString(), aggregateType, aggregateID, eventType, b)
	return err
}

func scanPolicies(rows pgx.Rows) ([]Policy, error) {
	var out []Policy
	for rows.Next() {
		var p Policy
		var risk, lines []byte
		var issuedAt *time.Time
		if err := rows.Scan(&p.ID, &p.TenantID, &p.PolicyNumber, &p.QuoteID, &p.PartyID, &p.ProductCode, &p.Status, &risk, &lines, &p.TotalMinor, &p.Currency, &p.EffectiveFrom, &p.EffectiveTo, &issuedAt, &p.InvoiceID); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(risk, &p.Risk)
		_ = json.Unmarshal(lines, &p.Lines)
		p.IssuedAt = issuedAt
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanClaimRow(row pgx.Row) (*Claim, error) {
	var c Claim
	var photos []byte
	if err := row.Scan(&c.ID, &c.ClaimNumber, &c.TenantID, &c.PolicyID, &c.Status, &c.Track, &c.Description, &c.Latitude, &c.Longitude, &c.EstimatedAmountMinor, &c.SettlementMinor, &c.Currency, &photos, &c.CreatedAt, &c.SettledAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	_ = json.Unmarshal(photos, &c.PhotoObjectKeys)
	return &c, nil
}

func scanClaimRows(rows pgx.Rows) (*Claim, error) {
	var c Claim
	var photos []byte
	if err := rows.Scan(&c.ID, &c.ClaimNumber, &c.TenantID, &c.PolicyID, &c.Status, &c.Track, &c.Description, &c.Latitude, &c.Longitude, &c.EstimatedAmountMinor, &c.SettlementMinor, &c.Currency, &photos, &c.CreatedAt, &c.SettledAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(photos, &c.PhotoObjectKeys)
	return &c, nil
}

func OpenRepository(ctx context.Context) (Repository, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return NewMemoryRepository(), nil
	}
	schema := os.Getenv("PG_SCHEMA")
	if schema == "" {
		schema = "pc_medhen"
	}
	return NewPostgres(ctx, dsn, schema)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
