package migrate_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/migrate"
)

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	ctx := context.Background()
	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("medhen"), tcpostgres.WithUsername("postgres"), tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(30*time.Second)))
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	conn, _ := container.ConnectionString(ctx, "sslmode=disable")
	pool, err := pgxpool.New(ctx, conn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestApply_IdempotentAndOrdered(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()

	migs := []migrate.Migration{
		{Version: 2, Name: "b", SQL: `CREATE TABLE b (id int)`},
		{Version: 1, Name: "a", SQL: `CREATE TABLE a (id int); CREATE INDEX ON a (id)`},
	}
	n, err := migrate.Apply(ctx, pool, migs)
	if err != nil || n != 2 {
		t.Fatalf("first apply = %d (%v), want 2", n, err)
	}

	// Re-apply is a no-op.
	n, err = migrate.Apply(ctx, pool, migs)
	if err != nil || n != 0 {
		t.Fatalf("re-apply = %d (%v), want 0", n, err)
	}

	// Adding a migration applies only the new one.
	migs = append(migs, migrate.Migration{Version: 3, Name: "c", SQL: `CREATE TABLE c (id int)`})
	n, err = migrate.Apply(ctx, pool, migs)
	if err != nil || n != 1 {
		t.Fatalf("incremental apply = %d (%v), want 1", n, err)
	}

	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM schema_migrations`).Scan(&count); err != nil || count != 3 {
		t.Fatalf("schema_migrations rows = %d (%v), want 3", count, err)
	}
	// All tables exist.
	for _, tbl := range []string{"a", "b", "c"} {
		var exists bool
		_ = pool.QueryRow(ctx, `SELECT to_regclass($1) IS NOT NULL`, tbl).Scan(&exists)
		if !exists {
			t.Fatalf("table %q was not created", tbl)
		}
	}
}

func TestApply_RollsBackFailedMigration(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()

	// The bad migration creates a table then errors — the table must not persist.
	migs := []migrate.Migration{
		{Version: 1, Name: "bad", SQL: `CREATE TABLE keep (id int); SELECT 1/0`},
	}
	if _, err := migrate.Apply(ctx, pool, migs); err == nil {
		t.Fatal("expected the failing migration to error")
	}
	var exists bool
	_ = pool.QueryRow(ctx, `SELECT to_regclass('keep') IS NOT NULL`).Scan(&exists)
	if exists {
		t.Fatal("failed migration must roll back its DDL")
	}
}
