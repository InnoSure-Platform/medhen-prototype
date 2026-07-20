package adapters_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/document/adapters"
	docapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/document/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func newService(t *testing.T) (*docapp.Service, *database.DB) {
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
	return docapp.NewService(db, adapters.NewDocumentRepository(db)), db
}

func TestGenerateCertificate_ContentAndIdempotency(t *testing.T) {
	svc, db := newService(t)
	ctx := context.Background()

	doc, err := svc.GenerateCertificate(ctx, "eic", "pol-1", "EIC/MOT/2026/000001", "Abebe Bikila")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if doc.Number != "COI-EIC/MOT/2026/000001" {
		t.Fatalf("unexpected COI number: %s", doc.Number)
	}
	if !strings.Contains(doc.Content, "Abebe Bikila") || !strings.Contains(doc.Content, "EIC/MOT/2026/000001") {
		t.Fatalf("COI content missing insured/policy: %q", doc.Content)
	}

	// Idempotent: re-generating for the same policy returns the same document and
	// does not emit a second event.
	again, _ := svc.GenerateCertificate(ctx, "eic", "pol-1", "EIC/MOT/2026/000001", "Abebe Bikila")
	if again.ID != doc.ID {
		t.Fatalf("idempotency broken: %s != %s", again.ID, doc.ID)
	}
	var n int
	_ = db.Pool().QueryRow(ctx, `SELECT count(*) FROM outbox WHERE topic='document.generated'`).Scan(&n)
	if n != 1 {
		t.Fatalf("expected 1 document.generated event, got %d", n)
	}
}
