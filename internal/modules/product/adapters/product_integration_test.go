package adapters_test

import (
	"context"
	"testing"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/adapters"
	ratingdomain "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/domain"
	ratingports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/rating/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/shopspring/decimal"
)

func newRepo(t *testing.T) *adapters.ProductRepository {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	ctx := context.Background()
	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("medhen"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	conn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}
	db, err := database.Connect(ctx, conn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(db.Close)

	if _, err := db.Pool().Exec(ctx, adapters.Schema); err != nil {
		t.Fatalf("schema: %v", err)
	}
	repo := adapters.NewProductRepository(db)
	if err := repo.Upsert(ctx, adapters.MotorProduct()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return repo
}

func TestRepo_GetSeededMotorProduct(t *testing.T) {
	repo := newRepo(t)
	p, err := repo.Get(context.Background(), "MOT")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(p.Coverages) != 2 || p.NameAmharic != "የተሽከርካሪ መድን" {
		t.Fatalf("unexpected product: %+v", p)
	}
}

func TestRepo_BaseRateAndFactor(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	base, ver, err := repo.BaseRate(ctx, "MOT", "OD")
	if err != nil || base.Minor() != 120000 || ver != "MOTOR-2026.1" {
		t.Fatalf("base rate = %v (%s, %v), want 1200.00/MOTOR-2026.1", base, ver, err)
	}
	f, _, err := repo.Factor(ctx, "MOT", "OD", "AGE", "young")
	if err != nil || !f.Equal(decimal.NewFromFloat(1.25)) {
		t.Fatalf("young factor = %v (%v), want 1.25", f, err)
	}
}

func TestRepo_Idempotency_ReSeed(t *testing.T) {
	repo := newRepo(t)
	// Upsert again should not error or duplicate.
	if err := repo.Upsert(context.Background(), adapters.MotorProduct()); err != nil {
		t.Fatalf("re-seed: %v", err)
	}
	ps, err := repo.List(context.Background())
	if err != nil || len(ps) != 1 {
		t.Fatalf("expected 1 product after re-seed, got %d (%v)", len(ps), err)
	}
}

// Cross-module: the rating engine, fed by the product-backed provider, prices a
// young driver's own-damage cover using the seeded 1.25 factor (1200 → 1500).
func TestRatingEngineUsesProductProvider(t *testing.T) {
	repo := newRepo(t)
	provider := adapters.NewRateProvider(repo)

	engine := ratingdomain.NewEngine(provider, ratingdomain.TaxPolicy{
		VATRate: decimal.NewFromFloat(0.15),
	})
	bd, err := engine.Calculate(context.Background(), ratingports.PremiumRequest{
		TenantID: "eic", ProductCode: "MOT", Coverages: []string{"OD"},
		RiskDimensions: map[string]string{"age_band": "young"},
	})
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	if bd.NetPremium.Minor() != 150000 {
		t.Fatalf("young OD net = %d, want 150000 (product-backed rating)", bd.NetPremium.Minor())
	}
}
