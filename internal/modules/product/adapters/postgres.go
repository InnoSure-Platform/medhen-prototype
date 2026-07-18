// Package adapters holds the product module's driven adapters (Postgres) and the
// rate-provider that satisfies the rating module's RateTableProvider port.
package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// Schema is the DDL for the product module's tables.
const Schema = `
CREATE TABLE IF NOT EXISTS products (
    code         TEXT PRIMARY KEY,
    lob          TEXT NOT NULL,
    name         TEXT NOT NULL,
    name_amharic TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL,
    rate_version TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS product_coverages (
    product_code TEXT NOT NULL REFERENCES products(code),
    code         TEXT NOT NULL,
    name         TEXT NOT NULL,
    name_amharic TEXT NOT NULL DEFAULT '',
    base_minor   BIGINT NOT NULL,
    PRIMARY KEY (product_code, code)
);
CREATE TABLE IF NOT EXISTS product_factors (
    product_code  TEXT NOT NULL,
    coverage_code TEXT NOT NULL,
    factor_type   TEXT NOT NULL,
    dimension     TEXT NOT NULL,
    factor        TEXT NOT NULL,
    PRIMARY KEY (product_code, coverage_code, factor_type, dimension)
);
`

// ProductRepository implements app.Repository over Postgres.
type ProductRepository struct{ db *database.DB }

// NewProductRepository builds the repository.
func NewProductRepository(db *database.DB) *ProductRepository { return &ProductRepository{db: db} }

var _ app.Repository = (*ProductRepository)(nil)

// Upsert writes a product with its coverages and factors in one transaction.
func (r *ProductRepository) Upsert(ctx context.Context, p *domain.Product) error {
	return r.db.WithinTx(ctx, func(ctx context.Context) error {
		q := r.db.Conn(ctx)
		if _, err := q.Exec(ctx,
			`INSERT INTO products (code, lob, name, name_amharic, status, rate_version)
			 VALUES ($1,$2,$3,$4,$5,$6)
			 ON CONFLICT (code) DO UPDATE SET lob=EXCLUDED.lob, name=EXCLUDED.name,
			   name_amharic=EXCLUDED.name_amharic, status=EXCLUDED.status,
			   rate_version=EXCLUDED.rate_version`,
			p.Code, p.LOB, p.Name, p.NameAmharic, p.Status, p.RateVersion); err != nil {
			return fmt.Errorf("product repo: upsert product: %w", err)
		}
		for _, c := range p.Coverages {
			if _, err := q.Exec(ctx,
				`INSERT INTO product_coverages (product_code, code, name, name_amharic, base_minor)
				 VALUES ($1,$2,$3,$4,$5)
				 ON CONFLICT (product_code, code) DO UPDATE SET name=EXCLUDED.name,
				   name_amharic=EXCLUDED.name_amharic, base_minor=EXCLUDED.base_minor`,
				p.Code, c.Code, c.Name, c.NameAmharic, c.BaseRate.Minor()); err != nil {
				return fmt.Errorf("product repo: upsert coverage: %w", err)
			}
		}
		for _, f := range p.Factors {
			if _, err := q.Exec(ctx,
				`INSERT INTO product_factors (product_code, coverage_code, factor_type, dimension, factor)
				 VALUES ($1,$2,$3,$4,$5)
				 ON CONFLICT (product_code, coverage_code, factor_type, dimension)
				   DO UPDATE SET factor=EXCLUDED.factor`,
				p.Code, f.CoverageCode, f.FactorType, f.Dimension, f.Value.String()); err != nil {
				return fmt.Errorf("product repo: upsert factor: %w", err)
			}
		}
		return nil
	})
}

// Get loads a product with its coverages.
func (r *ProductRepository) Get(ctx context.Context, code string) (*domain.Product, error) {
	q := r.db.Conn(ctx)
	var p domain.Product
	err := q.QueryRow(ctx,
		`SELECT code, lob, name, name_amharic, status, rate_version FROM products WHERE code=$1`, code).
		Scan(&p.Code, &p.LOB, &p.Name, &p.NameAmharic, &p.Status, &p.RateVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, app.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("product repo: get: %w", err)
	}

	rows, err := q.Query(ctx,
		`SELECT code, name, name_amharic, base_minor FROM product_coverages WHERE product_code=$1 ORDER BY code`, code)
	if err != nil {
		return nil, fmt.Errorf("product repo: coverages: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var c domain.Coverage
		var minor int64
		if err := rows.Scan(&c.Code, &c.Name, &c.NameAmharic, &minor); err != nil {
			return nil, err
		}
		c.BaseRate = money.FromMinor(minor)
		p.Coverages = append(p.Coverages, c)
	}
	return &p, rows.Err()
}

// List returns all products (without factors).
func (r *ProductRepository) List(ctx context.Context) ([]*domain.Product, error) {
	rows, err := r.db.Conn(ctx).Query(ctx, `SELECT code FROM products ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("product repo: list: %w", err)
	}
	var codes []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			rows.Close()
			return nil, err
		}
		codes = append(codes, c)
	}
	rows.Close()

	out := make([]*domain.Product, 0, len(codes))
	for _, c := range codes {
		p, err := r.Get(ctx, c)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// BaseRate returns the base premium and rate version for a coverage.
func (r *ProductRepository) BaseRate(ctx context.Context, productCode, coverageCode string) (money.Amount, string, error) {
	var minor int64
	var version string
	err := r.db.Conn(ctx).QueryRow(ctx,
		`SELECT c.base_minor, p.rate_version
		   FROM product_coverages c JOIN products p ON p.code=c.product_code
		  WHERE c.product_code=$1 AND c.code=$2`, productCode, coverageCode).
		Scan(&minor, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return money.Zero(), "", fmt.Errorf("no base rate for %s/%s", productCode, coverageCode)
	}
	if err != nil {
		return money.Zero(), "", fmt.Errorf("product repo: base rate: %w", err)
	}
	return money.FromMinor(minor), version, nil
}

// Factor returns a rating factor and version for a coverage/dimension.
func (r *ProductRepository) Factor(ctx context.Context, productCode, coverageCode, factorType, dimension string) (decimal.Decimal, string, error) {
	var factorStr, version string
	err := r.db.Conn(ctx).QueryRow(ctx,
		`SELECT f.factor, p.rate_version
		   FROM product_factors f JOIN products p ON p.code=f.product_code
		  WHERE f.product_code=$1 AND f.coverage_code=$2 AND f.factor_type=$3 AND f.dimension=$4`,
		productCode, coverageCode, factorType, dimension).Scan(&factorStr, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return decimal.Zero, "", fmt.Errorf("no %s factor for %s/%s dim=%q", factorType, productCode, coverageCode, dimension)
	}
	if err != nil {
		return decimal.Zero, "", fmt.Errorf("product repo: factor: %w", err)
	}
	d, err := decimal.NewFromString(factorStr)
	if err != nil {
		return decimal.Zero, "", fmt.Errorf("product repo: parse factor %q: %w", factorStr, err)
	}
	return d, version, nil
}
