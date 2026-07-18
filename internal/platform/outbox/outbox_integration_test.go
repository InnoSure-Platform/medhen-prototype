package outbox_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgres(t *testing.T) *database.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	ctx := context.Background()
	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
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

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}
	db, err := database.Connect(ctx, connStr)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(db.Close)

	if _, err := db.Pool().Exec(ctx, outbox.Schema); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	return db
}

// recorder is a thread-safe Publisher that records delivered message IDs.
type recorder struct {
	mu  sync.Mutex
	ids []string
}

func (r *recorder) Publish(_ context.Context, m outbox.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ids = append(r.ids, m.ID)
	return nil
}

func (r *recorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.ids)
}

func msg() outbox.Message {
	return outbox.Message{
		ID: ids.New(), Topic: "policy.bound", AggregateType: "policy",
		AggregateID: ids.New(), Payload: []byte(`{"ok":true}`),
	}
}

func TestWithinTx_CommitPersistsOutbox(t *testing.T) {
	db := startPostgres(t)
	ctx := context.Background()

	m := msg()
	err := db.WithinTx(ctx, func(ctx context.Context) error {
		return outbox.Write(ctx, db.Conn(ctx), m)
	})
	if err != nil {
		t.Fatalf("within tx: %v", err)
	}

	var n int
	if err := db.Pool().QueryRow(ctx, `SELECT count(*) FROM outbox WHERE id=$1`, m.ID).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 committed outbox row, got %d", n)
	}
}

func TestWithinTx_RollbackDiscardsOutbox(t *testing.T) {
	db := startPostgres(t)
	ctx := context.Background()

	m := msg()
	wantErr := fmt.Errorf("boom")
	err := db.WithinTx(ctx, func(ctx context.Context) error {
		if e := outbox.Write(ctx, db.Conn(ctx), m); e != nil {
			return e
		}
		return wantErr // force rollback
	})
	if err == nil {
		t.Fatal("expected error to trigger rollback")
	}

	var n int
	_ = db.Pool().QueryRow(ctx, `SELECT count(*) FROM outbox WHERE id=$1`, m.ID).Scan(&n)
	if n != 0 {
		t.Fatalf("rollback should discard outbox row, found %d", n)
	}
}

func TestRelay_PublishesAndMarksProcessed(t *testing.T) {
	db := startPostgres(t)
	ctx := context.Background()

	m := msg()
	if err := db.WithinTx(ctx, func(ctx context.Context) error {
		return outbox.Write(ctx, db.Conn(ctx), m)
	}); err != nil {
		t.Fatalf("write: %v", err)
	}

	rec := &recorder{}
	relay := outbox.NewRelay(db, rec, 100, nil)

	n, err := relay.ProcessBatch(ctx)
	if err != nil || n != 1 {
		t.Fatalf("ProcessBatch = %d (%v), want 1", n, err)
	}
	// Idempotent: nothing left to process.
	n2, _ := relay.ProcessBatch(ctx)
	if n2 != 0 {
		t.Fatalf("second ProcessBatch = %d, want 0 (already processed)", n2)
	}
	if rec.count() != 1 {
		t.Fatalf("published %d, want 1", rec.count())
	}
}

// The core C7 regression test: two relay workers draining concurrently must not
// double-publish any message.
func TestRelay_ConcurrentWorkersNoDoublePublish(t *testing.T) {
	db := startPostgres(t)
	ctx := context.Background()

	const total = 200
	for i := 0; i < total; i++ {
		if err := db.WithinTx(ctx, func(ctx context.Context) error {
			return outbox.Write(ctx, db.Conn(ctx), msg())
		}); err != nil {
			t.Fatalf("seed write: %v", err)
		}
	}

	rec := &recorder{}
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			relay := outbox.NewRelay(db, rec, 10, nil)
			for {
				n, err := relay.ProcessBatch(ctx)
				if err != nil {
					t.Errorf("worker batch: %v", err)
					return
				}
				if n == 0 {
					return
				}
			}
		}()
	}
	wg.Wait()

	// Exactly `total` deliveries, all unique.
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.ids) != total {
		t.Fatalf("delivered %d messages, want %d (double-publish or loss)", len(rec.ids), total)
	}
	seen := make(map[string]bool, total)
	for _, id := range rec.ids {
		if seen[id] {
			t.Fatalf("message %s published more than once", id)
		}
		seen[id] = true
	}
}
