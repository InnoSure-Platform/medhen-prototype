package database_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestRLS_TenantIsolation proves the Phase-5 security model end-to-end: a table
// with a tenant_isolation RLS policy, a least-privilege app role, and
// WithinTenantTx setting app.current_tenant — the app role sees only its tenant's
// rows and cannot write rows for another tenant.
func TestRLS_TenantIsolation(t *testing.T) {
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

	ownerURL, _ := container.ConnectionString(ctx, "sslmode=disable")
	owner, err := database.Connect(ctx, ownerURL)
	if err != nil {
		t.Fatalf("owner connect: %v", err)
	}
	t.Cleanup(owner.Close)

	// Set up a tenant-scoped table with RLS + a least-privilege role (as owner).
	setup := []string{
		`CREATE TABLE things (id text PRIMARY KEY, tenant_id text NOT NULL, v text)`,
		`ALTER TABLE things ENABLE ROW LEVEL SECURITY`,
		`CREATE POLICY tenant_isolation ON things
		   USING (tenant_id = current_setting('app.current_tenant', true))
		   WITH CHECK (tenant_id = current_setting('app.current_tenant', true))`,
		`DO $$ BEGIN IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname='medhen_app')
		   THEN CREATE ROLE medhen_app LOGIN PASSWORD 'medhen_app'; END IF; END $$`,
		`GRANT SELECT, INSERT, UPDATE, DELETE ON things TO medhen_app`,
		// Owner bypasses RLS, so it can seed both tenants' rows.
		`INSERT INTO things VALUES ('1','eic','a1'), ('2','awash','b1')`,
	}
	for _, s := range setup {
		if _, err := owner.Pool().Exec(ctx, s); err != nil {
			t.Fatalf("setup %q: %v", s, err)
		}
	}

	// Connect as the least-privilege app role (subject to RLS).
	appURL := swapUser(t, ownerURL, "medhen_app", "medhen_app")
	app, err := database.Connect(ctx, appURL)
	if err != nil {
		t.Fatalf("app connect: %v", err)
	}
	t.Cleanup(app.Close)

	// Fail-closed: with no tenant set, the app role sees nothing.
	if got := count(t, app, ctx); got != 0 {
		t.Fatalf("no-tenant read saw %d rows, want 0 (RLS fail-closed)", got)
	}

	// Scoped read: each tenant sees exactly its own row.
	for tenant, want := range map[string]string{"eic": "a1", "awash": "b1"} {
		if err := app.WithinTenantTx(ctx, tenant, func(ctx context.Context) error {
			var n int
			if err := app.Conn(ctx).QueryRow(ctx, `SELECT count(*) FROM things`).Scan(&n); err != nil {
				return err
			}
			if n != 1 {
				t.Fatalf("tenant %s saw %d rows, want 1", tenant, n)
			}
			var v string
			if err := app.Conn(ctx).QueryRow(ctx, `SELECT v FROM things`).Scan(&v); err != nil {
				return err
			}
			if v != want {
				t.Fatalf("tenant %s saw %q, want %q", tenant, v, want)
			}
			return nil
		}); err != nil {
			t.Fatalf("tenant %s read: %v", tenant, err)
		}
	}

	// WITH CHECK: the app role cannot insert a row for a different tenant.
	err = app.WithinTenantTx(ctx, "eic", func(ctx context.Context) error {
		_, e := app.Conn(ctx).Exec(ctx, `INSERT INTO things VALUES ('9','awash','x')`)
		return e
	})
	if err == nil {
		t.Fatal("cross-tenant insert must be blocked by the RLS WITH CHECK")
	}

	// A same-tenant insert succeeds.
	if err := app.WithinTenantTx(ctx, "eic", func(ctx context.Context) error {
		_, e := app.Conn(ctx).Exec(ctx, `INSERT INTO things VALUES ('10','eic','ok')`)
		return e
	}); err != nil {
		t.Fatalf("same-tenant insert should succeed: %v", err)
	}
}

func count(t *testing.T, db *database.DB, ctx context.Context) int {
	t.Helper()
	var n int
	if err := db.Pool().QueryRow(ctx, `SELECT count(*) FROM things`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func swapUser(t *testing.T, rawURL, user, pass string) string {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	u.User = url.UserPassword(user, pass)
	return u.String()
}
