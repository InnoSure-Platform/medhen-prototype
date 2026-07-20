// Package database provides the shared Postgres pool and a Unit-of-Work.
//
// The Unit-of-Work makes a transaction ambient on the context: WithinTx opens a
// pgx.Tx and stashes it on the context; repositories call Conn(ctx) to get the
// ambient tx (or the pool when none is active). This lets a use case wrap
// several repository writes — plus the outbox insert — in one atomic commit
// without repositories knowing about transaction boundaries.
package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier is the subset of pgx behaviour shared by *pgxpool.Pool and pgx.Tx.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txKey struct{}
type tenantKey struct{}

// WithTenant binds a tenant to the context. When a transaction later begins on
// this context (WithinTx), the tenant is applied as the app.current_tenant
// setting so Postgres row-level-security policies scope every statement to it.
func WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenantID)
}

// TenantFromContext returns the bound tenant, or "" if none.
func TenantFromContext(ctx context.Context) string {
	t, _ := ctx.Value(tenantKey{}).(string)
	return t
}

// DB owns the connection pool.
type DB struct {
	pool *pgxpool.Pool
}

// Connect opens and verifies a pool against the given URL.
func Connect(ctx context.Context, url string) (*DB, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("database: new pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}
	return &DB{pool: pool}, nil
}

// FromPool wraps an existing pool (used by tests).
func FromPool(pool *pgxpool.Pool) *DB { return &DB{pool: pool} }

// Pool exposes the underlying pool for migrations/health.
func (d *DB) Pool() *pgxpool.Pool { return d.pool }

// Health pings the database.
func (d *DB) Health(ctx context.Context) error { return d.pool.Ping(ctx) }

// Close releases the pool.
func (d *DB) Close() { d.pool.Close() }

// Conn returns the ambient transaction if one is active on ctx, otherwise the
// pool. Repositories should always use this rather than capturing the pool.
func (d *DB) Conn(ctx context.Context) Querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return d.pool
}

// WithinTx runs fn inside a transaction. If a transaction is already active on
// ctx, fn joins it (no nested BEGIN) so composed use cases share one commit.
// The tx commits when fn returns nil and rolls back otherwise (including panic).
func (d *DB) WithinTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx) // already in a UoW — participate
	}

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("database: begin: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	txCtx := context.WithValue(ctx, txKey{}, tx)

	// Bind the tenant to the transaction for row-level security. set_config(...,
	// true) scopes it to this tx. When no tenant is bound (background workers,
	// dev), the setting is left empty — safe under the owner/system role, which
	// bypasses RLS.
	if tenant := TenantFromContext(ctx); tenant != "" {
		if _, cfgErr := tx.Exec(ctx, `SELECT set_config('app.current_tenant', $1, true)`, tenant); cfgErr != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("database: set tenant: %w", cfgErr)
		}
	}

	if err = fn(txCtx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			return errors.Join(err, fmt.Errorf("rollback: %w", rbErr))
		}
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("database: commit: %w", err)
	}
	return nil
}

// WithinTenantTx runs fn in a transaction with the tenant bound to the
// `app.current_tenant` setting for the life of the transaction. Postgres
// row-level-security policies read this setting to enforce tenant isolation, so
// tenant-scoped writes/reads made under the least-privilege application role only
// see their own tenant's rows. Passing an empty tenantID is rejected.
func (d *DB) WithinTenantTx(ctx context.Context, tenantID string, fn func(ctx context.Context) error) error {
	if tenantID == "" {
		return errors.New("database: WithinTenantTx requires a non-empty tenant")
	}
	// WithinTx applies the bound tenant as app.current_tenant when it begins.
	return d.WithinTx(WithTenant(ctx, tenantID), fn)
}
