package idempotency_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/idempotency"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startRedis(t *testing.T) redis.UniversalClient {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	ctx := context.Background()
	container, err := tcredis.Run(ctx, "redis:7-alpine",
		testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections")))
	if err != nil {
		t.Fatalf("start redis: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}
	opts, err := redis.ParseURL(uri)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	client := redis.NewClient(opts)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestAcquireNewThenReplay(t *testing.T) {
	store := idempotency.New(startRedis(t), time.Minute)
	ctx := context.Background()

	r, err := store.Acquire(ctx, "req-1")
	if err != nil || r.Status != idempotency.StatusNew {
		t.Fatalf("first Acquire = %v (%v), want StatusNew", r.Status, err)
	}

	// Before completion, a duplicate sees in-progress.
	if r2, _ := store.Acquire(ctx, "req-1"); r2.Status != idempotency.StatusInProgress {
		t.Fatalf("duplicate before complete = %v, want StatusInProgress", r2.Status)
	}

	if err := store.Complete(ctx, "req-1", []byte(`{"policy":"EIC/MOT/2026/000001"}`)); err != nil {
		t.Fatalf("complete: %v", err)
	}

	// After completion, a duplicate replays the cached response.
	r3, _ := store.Acquire(ctx, "req-1")
	if r3.Status != idempotency.StatusDone {
		t.Fatalf("duplicate after complete = %v, want StatusDone", r3.Status)
	}
	if string(r3.Response) != `{"policy":"EIC/MOT/2026/000001"}` {
		t.Fatalf("replay response = %q", string(r3.Response))
	}
}

// Only ONE of many concurrent claimants may win — the core H7 regression test.
func TestConcurrentAcquireSingleWinner(t *testing.T) {
	store := idempotency.New(startRedis(t), time.Minute)
	ctx := context.Background()

	const n = 100
	var winners int64
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if r, err := store.Acquire(ctx, "hot-key"); err == nil && r.Status == idempotency.StatusNew {
				atomic.AddInt64(&winners, 1)
			}
		}()
	}
	wg.Wait()

	if winners != 1 {
		t.Fatalf("StatusNew winners = %d, want exactly 1 (atomic SETNX)", winners)
	}
}

func TestReleaseAllowsRetry(t *testing.T) {
	store := idempotency.New(startRedis(t), time.Minute)
	ctx := context.Background()

	if r, _ := store.Acquire(ctx, "k"); r.Status != idempotency.StatusNew {
		t.Fatal("want StatusNew")
	}
	if err := store.Release(ctx, "k"); err != nil {
		t.Fatalf("release: %v", err)
	}
	if r, _ := store.Acquire(ctx, "k"); r.Status != idempotency.StatusNew {
		t.Fatal("after release, want StatusNew again")
	}
}
