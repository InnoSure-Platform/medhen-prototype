package adapters_test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/reporting/adapters"
	reportapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/reporting/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
)

func newService(t *testing.T) *reportapp.Service {
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
	return reportapp.NewService(adapters.NewKPIRepository(db))
}

func TestKPIs_ComputesRealLossRatio(t *testing.T) {
	svc := newService(t)
	ctx := context.Background()

	// Two policies: 2680 + 2680 = 5360 premium; one claim paid 1340.
	if err := svc.RecordPolicy(ctx, "eic", 268000); err != nil {
		t.Fatalf("record policy: %v", err)
	}
	_ = svc.RecordPolicy(ctx, "eic", 268000)
	_ = svc.RecordClaim(ctx, "eic", 134000)

	view, err := svc.KPIs(ctx, "eic")
	if err != nil {
		t.Fatalf("kpis: %v", err)
	}
	if view.PolicyCount != 2 || view.ClaimCount != 1 {
		t.Fatalf("counts = %d policies / %d claims, want 2/1", view.PolicyCount, view.ClaimCount)
	}
	if view.PremiumWrittenMinor != 536000 || view.ClaimsPaidMinor != 134000 {
		t.Fatalf("premium=%d claims=%d, want 536000/134000", view.PremiumWrittenMinor, view.ClaimsPaidMinor)
	}
	// loss ratio = 134000/536000 = 0.25 (real, not the pre-refactor dummy value).
	if view.LossRatio != 0.25 {
		t.Fatalf("loss ratio = %v, want 0.25", view.LossRatio)
	}
	if view.CombinedRatio != 0.55 { // 0.25 + 0.30 assumed expense
		t.Fatalf("combined ratio = %v, want 0.55", view.CombinedRatio)
	}
}

func TestKPIs_EmptyTenantIsZero(t *testing.T) {
	view, err := newService(t).KPIs(context.Background(), "nobody")
	if err != nil {
		t.Fatalf("kpis: %v", err)
	}
	if view.LossRatio != 0 || view.PremiumWrittenMinor != 0 {
		t.Fatalf("expected zero KPIs for empty tenant, got %+v", view)
	}
}
