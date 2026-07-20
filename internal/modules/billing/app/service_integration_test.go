package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/adapters"
	billingapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func newService(t *testing.T) (*billingapp.Service, *database.DB) {
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
	svc := billingapp.NewService(billingapp.Deps{
		DB: db, Invoices: adapters.NewInvoiceRepository(db), Payments: adapters.NewPaymentRepository(db),
	})
	return svc, db
}

func outboxCount(t *testing.T, db *database.DB, topic string) int {
	t.Helper()
	var n int
	if err := db.Pool().QueryRow(context.Background(),
		`SELECT count(*) FROM outbox WHERE topic=$1`, topic).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func TestRaiseInvoiceForPolicy_IsIdempotent(t *testing.T) {
	svc, db := newService(t)
	ctx := context.Background()

	inv1, err := svc.RaiseInvoiceForPolicy(ctx, "eic", "pol-1", "party-1", 268000)
	if err != nil {
		t.Fatalf("raise: %v", err)
	}
	if inv1.AmountDue.Minor() != 268000 || inv1.Status != domain.InvoiceOpen {
		t.Fatalf("unexpected invoice: %+v", inv1)
	}

	// Redelivery of policy.issued must not create a second invoice.
	inv2, err := svc.RaiseInvoiceForPolicy(ctx, "eic", "pol-1", "party-1", 268000)
	if err != nil {
		t.Fatalf("re-raise: %v", err)
	}
	if inv2.ID != inv1.ID {
		t.Fatalf("idempotency broken: %s != %s", inv2.ID, inv1.ID)
	}
	if n := outboxCount(t, db, domain.TopicInvoiceRaised); n != 1 {
		t.Fatalf("expected 1 invoice.raised event, got %d", n)
	}
}

func TestRecordPayment_FullSettlement(t *testing.T) {
	svc, db := newService(t)
	ctx := context.Background()

	inv, _ := svc.RaiseInvoiceForPolicy(ctx, "eic", "pol-2", "party-1", 268000)

	updated, err := svc.RecordPayment(ctx, "eic", inv.ID, money.FromMinor(268000), "TELEBIRR", "TB-REF-1")
	if err != nil {
		t.Fatalf("record payment: %v", err)
	}
	if updated.Status != domain.InvoicePaid || !updated.Outstanding().IsZero() {
		t.Fatalf("expected PAID with zero outstanding, got %+v", updated)
	}
	if n := outboxCount(t, db, domain.TopicPaymentReceived); n != 1 {
		t.Fatalf("expected 1 payment.received event, got %d", n)
	}
}

func TestRecordPayment_PartialThenFull(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()

	inv, _ := svc.RaiseInvoiceForPolicy(ctx, "eic", "pol-3", "party-1", 268000)

	partial, _ := svc.RecordPayment(ctx, "eic", inv.ID, money.FromMinor(100000), "TELEBIRR", "TB-1")
	if partial.Status != domain.InvoicePartiallyPaid || partial.Outstanding().Minor() != 168000 {
		t.Fatalf("expected PARTIALLY_PAID with 168000 outstanding, got %+v", partial)
	}
	full, _ := svc.RecordPayment(ctx, "eic", inv.ID, money.FromMinor(168000), "TELEBIRR", "TB-2")
	if full.Status != domain.InvoicePaid {
		t.Fatalf("expected PAID after second payment, got %s", full.Status)
	}
}

func TestRecordPayment_RejectsNonPositive(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	inv, _ := svc.RaiseInvoiceForPolicy(ctx, "eic", "pol-4", "party-1", 100000)
	if _, err := svc.RecordPayment(ctx, "eic", inv.ID, money.Zero(), "TELEBIRR", "x"); err != domain.ErrNonPositivePayment {
		t.Fatalf("expected ErrNonPositivePayment, got %v", err)
	}
}
