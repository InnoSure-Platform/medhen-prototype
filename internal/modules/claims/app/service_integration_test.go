package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/adapters"
	claimsapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/claims/domain"
	policyports "github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// stubPolicyReader lets the claims tests control policy lookups.
type stubPolicyReader struct {
	view policyports.PolicyView
	err  error
}

func (s stubPolicyReader) GetPolicy(context.Context, string, string) (policyports.PolicyView, error) {
	return s.view, s.err
}

func newService(t *testing.T, reader policyports.Reader) (*claimsapp.Service, *database.DB) {
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

	for _, ddl := range []string{outbox.Schema, adapters.Schema} {
		if _, err := db.Pool().Exec(ctx, ddl); err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	svc := claimsapp.NewService(claimsapp.Deps{
		DB: db, Claims: adapters.NewClaimRepository(db), Policy: reader, FastTrackLimit: money.FromInt(50000),
	})
	return svc, db
}

func issuedReader() stubPolicyReader {
	return stubPolicyReader{view: policyports.PolicyView{ID: "pol-1", TenantID: "eic", PartyID: "party-1", Status: "ISSUED"}}
}

func outboxCount(t *testing.T, db *database.DB, topic string) int {
	t.Helper()
	var n int
	if err := db.Pool().QueryRow(context.Background(), `SELECT count(*) FROM outbox WHERE topic=$1`, topic).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func TestFileFNOL_ValidatesPolicyAndEmitsEvent(t *testing.T) {
	svc, db := newService(t, issuedReader())
	ctx := context.Background()

	c, err := svc.FileFNOL(ctx, claimsapp.FNOLInput{
		PolicyID: "pol-1", TenantID: "eic", Description: "rear-end collision",
		Latitude: 9.03, Longitude: 38.74, ReserveMinor: 4000000,
	})
	if err != nil {
		t.Fatalf("fnol: %v", err)
	}
	if c.Status != domain.StatusFiled || c.PartyID != "party-1" {
		t.Fatalf("unexpected claim: %+v", c)
	}
	if outboxCount(t, db, domain.TopicClaimFiled) != 1 {
		t.Fatal("expected 1 claims.filed event")
	}
}

func TestFileFNOL_RejectsMissingOrInactivePolicy(t *testing.T) {
	// missing policy
	svc, _ := newService(t, stubPolicyReader{err: context.DeadlineExceeded})
	if _, err := svc.FileFNOL(context.Background(), claimsapp.FNOLInput{PolicyID: "x", TenantID: "eic"}); err != claimsapp.ErrPolicyNotFound {
		t.Fatalf("expected ErrPolicyNotFound, got %v", err)
	}

	// cancelled policy
	svc2, _ := newService(t, stubPolicyReader{view: policyports.PolicyView{Status: "CANCELLED"}})
	if _, err := svc2.FileFNOL(context.Background(), claimsapp.FNOLInput{PolicyID: "x", TenantID: "eic"}); err != claimsapp.ErrPolicyNotActive {
		t.Fatalf("expected ErrPolicyNotActive, got %v", err)
	}
}

func TestFastTrackSettle_WithinAuthority(t *testing.T) {
	svc, db := newService(t, issuedReader())
	ctx := context.Background()
	c, _ := svc.FileFNOL(ctx, claimsapp.FNOLInput{PolicyID: "pol-1", TenantID: "eic", ReserveMinor: 4000000})

	settled, err := svc.FastTrackSettle(ctx, "eic", c.ID, money.FromInt(30000)) // < 50000 limit
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if settled.Status != domain.StatusSettled || settled.SettledAmount.Minor() != 3000000 {
		t.Fatalf("unexpected settlement: %+v", settled)
	}
	if outboxCount(t, db, domain.TopicClaimSettled) != 1 {
		t.Fatal("expected 1 claims.settled event")
	}
}

func TestFastTrackSettle_OverAuthorityRefersAtomically(t *testing.T) {
	svc, db := newService(t, issuedReader())
	ctx := context.Background()
	c, _ := svc.FileFNOL(ctx, claimsapp.FNOLInput{PolicyID: "pol-1", TenantID: "eic", ReserveMinor: 9000000})

	if _, err := svc.FastTrackSettle(ctx, "eic", c.ID, money.FromInt(80000)); err != domain.ErrAuthorityExceeded {
		t.Fatalf("expected ErrAuthorityExceeded, got %v", err)
	}
	// Atomic: no settled event, claim still FILED.
	if outboxCount(t, db, domain.TopicClaimSettled) != 0 {
		t.Fatal("over-authority settle must not emit an event")
	}
	got, _ := svc.GetClaim(ctx, "eic", c.ID)
	if got.Status != domain.StatusFiled {
		t.Fatalf("claim should remain FILED after refer, got %s", got.Status)
	}
}
