// Package migrate is a small, transactional schema migrator. Each migration runs
// in its own transaction and is recorded in schema_migrations, so Apply is
// idempotent and safe to run at every boot.
//
// It is deliberately module-agnostic (the platform kernel must not import
// bounded-context modules). The composition root assembles the ordered migration
// list — reusing each module's own DDL — and calls Apply. This keeps a single
// source of truth for schema (the module `Schema` consts) with no drift.
package migrate

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migration is one ordered, versioned schema change. SQL may contain multiple
// statements (executed via the simple protocol).
type Migration struct {
	Version int64
	Name    string
	SQL     string
}

const trackingDDL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    BIGINT PRIMARY KEY,
    name       TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`

// Apply runs every not-yet-applied migration in version order, each in its own
// transaction, recording it in schema_migrations. It returns the number applied.
func Apply(ctx context.Context, pool *pgxpool.Pool, migrations []Migration) (int, error) {
	if _, err := pool.Exec(ctx, trackingDDL); err != nil {
		return 0, fmt.Errorf("migrate: ensure tracking table: %w", err)
	}

	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return 0, err
	}

	ordered := append([]Migration(nil), migrations...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Version < ordered[j].Version })

	count := 0
	for _, m := range ordered {
		if applied[m.Version] {
			continue
		}
		if err := applyOne(ctx, pool, m); err != nil {
			return count, fmt.Errorf("migrate: %d_%s: %w", m.Version, m.Name, err)
		}
		count++
	}
	return count, nil
}

func appliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[int64]bool, error) {
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("migrate: load applied: %w", err)
	}
	defer rows.Close()
	applied := make(map[int64]bool)
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func applyOne(ctx context.Context, pool *pgxpool.Pool, m Migration) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit

	if _, err := tx.Exec(ctx, m.SQL); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, m.Version, m.Name); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
