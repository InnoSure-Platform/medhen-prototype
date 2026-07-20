package adapters_test

import (
	"context"
	"testing"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/audit/adapters"
	auditapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/audit/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func newRepo(t *testing.T) *adapters.AuditRepository {
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
	db, err := database.Connect(ctx, conn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(db.Close)
	if _, err := db.Pool().Exec(ctx, adapters.Schema); err != nil {
		t.Fatalf("schema: %v", err)
	}
	return adapters.NewAuditRepository(db)
}

func TestAppendAndListWithFilters(t *testing.T) {
	repo := newRepo(t)
	rec := auditapp.NewRecorder(repo)
	ctx := context.Background()

	if err := rec.Record(ctx, "policy.issued", []byte(`{"tenant_id":"eic","policy_id":"p1"}`)); err != nil {
		t.Fatalf("record: %v", err)
	}
	_ = rec.Record(ctx, "billing.invoice_raised", []byte(`{"tenant_id":"eic","invoice_id":"i1"}`))
	_ = rec.Record(ctx, "policy.issued", []byte(`{"tenant_id":"other","policy_id":"p2"}`))

	// Tenant-scoped list.
	all, err := rec.List(ctx, "eic", "", 100)
	if err != nil || len(all) != 2 {
		t.Fatalf("list eic = %d (%v), want 2", len(all), err)
	}
	// Topic filter.
	issued, _ := rec.List(ctx, "eic", "policy.issued", 100)
	if len(issued) != 1 || issued[0].Topic != "policy.issued" {
		t.Fatalf("topic filter returned %d, want 1", len(issued))
	}
	// Tenant isolation.
	other, _ := rec.List(ctx, "other", "", 100)
	if len(other) != 1 {
		t.Fatalf("tenant isolation: other = %d, want 1", len(other))
	}
}
