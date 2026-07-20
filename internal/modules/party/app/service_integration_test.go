package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/adapters"
	partyapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/party/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
)

func newService(t *testing.T) (*partyapp.Service, *database.DB) {
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

	for _, ddl := range []string{outbox.Schema, adapters.Schema} {
		if _, err := db.Pool().Exec(ctx, ddl); err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return partyapp.NewService(db, adapters.NewPartyRepository(db)), db
}

func validInput() partyapp.RegisterInput {
	return partyapp.RegisterInput{
		FullName: "Abebe Bikila", FullNameAmharic: "አበበ ቢቂላ",
		PhoneE164: "+251911000000", NationalID: "ETH-0001",
		Address: domain.Address{Region: "Addis Ababa", Zone: "Bole", Woreda: "03", Kebele: "05"},
	}
}

func outboxCount(t *testing.T, db *database.DB) int {
	t.Helper()
	var n int
	if err := db.Pool().QueryRow(context.Background(),
		`SELECT count(*) FROM outbox WHERE topic=$1`, domain.TopicPartyRegistered).Scan(&n); err != nil {
		t.Fatalf("count outbox: %v", err)
	}
	return n
}

func TestRegister_PersistsPartyAndOutboxAtomically(t *testing.T) {
	svc, db := newService(t)
	ctx := context.Background()

	in := validInput()
	in.TenantID = "eic"
	id, err := svc.Register(ctx, in)
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Party persisted.
	got, err := svc.Get(ctx, "eic", id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.FullNameAmharic != "አበበ ቢቂላ" {
		t.Fatalf("amharic name not persisted: %q", got.FullNameAmharic)
	}
	// Outbox event written in the same tx.
	if outboxCount(t, db) != 1 {
		t.Fatalf("expected 1 outbox event, got %d", outboxCount(t, db))
	}
}

func TestRegister_DuplicateNationalID_RollsBackBoth(t *testing.T) {
	svc, db := newService(t)
	ctx := context.Background()

	in := validInput()
	in.TenantID = "eic"
	if _, err := svc.Register(ctx, in); err != nil {
		t.Fatalf("first register: %v", err)
	}
	before := outboxCount(t, db)

	// Same national id → unique-index violation inside the tx.
	if _, err := svc.Register(ctx, in); err == nil {
		t.Fatal("expected duplicate national id to fail")
	}

	// Atomicity: the failed attempt must not have written an outbox event.
	if after := outboxCount(t, db); after != before {
		t.Fatalf("outbox events changed on failed register: before=%d after=%d", before, after)
	}
}

func TestRegister_ValidationError_NoWrite(t *testing.T) {
	svc, db := newService(t)
	ctx := context.Background()

	in := validInput()
	in.TenantID = "eic"
	in.FullName = "" // invalid
	if _, err := svc.Register(ctx, in); err == nil {
		t.Fatal("expected validation error")
	}
	if outboxCount(t, db) != 0 {
		t.Fatal("no outbox event should be written on validation failure")
	}
}

func TestGet_NotFound(t *testing.T) {
	svc, _ := newService(t)
	if _, err := svc.Get(context.Background(), "eic", "nope"); err != partyapp.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGet_TenantIsolation(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()

	in := validInput()
	in.TenantID = "eic"
	id, _ := svc.Register(ctx, in)

	// Another tenant must not see it.
	if _, err := svc.Get(ctx, "other", id); err != partyapp.ErrNotFound {
		t.Fatalf("cross-tenant read should be NotFound, got %v", err)
	}
}
