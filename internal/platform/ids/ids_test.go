package ids

import (
	"context"
	"strings"
	"sync"
	"testing"
)

func TestNewIsUniqueAndSortable(t *testing.T) {
	a := New()
	b := New()
	if a == b {
		t.Fatal("expected distinct ULIDs")
	}
	// ULIDs are 26 chars, Crockford base32, monotonic within the same ms.
	if len(a) != 26 {
		t.Fatalf("ULID length = %d, want 26", len(a))
	}
}

func TestNewWithPrefix(t *testing.T) {
	id := NewWithPrefix("pol")
	if !strings.HasPrefix(id, "pol_") {
		t.Fatalf("prefix missing: %s", id)
	}
}

func TestPolicyNumberFormat(t *testing.T) {
	got := PolicyNumber("EIC", "MOT", 2026, 42)
	if got != "EIC/MOT/2026/000042" {
		t.Fatalf("PolicyNumber = %q, want EIC/MOT/2026/000042", got)
	}
}

func TestInMemorySequencerMonotonic(t *testing.T) {
	s := NewInMemorySequencer()
	ctx := context.Background()
	for i := int64(1); i <= 5; i++ {
		n, err := s.Next(ctx, "MOT-2026")
		if err != nil || n != i {
			t.Fatalf("Next() = %d (%v), want %d", n, err, i)
		}
	}
	// Independent sequences don't interfere.
	if n, _ := s.Next(ctx, "PROP-2026"); n != 1 {
		t.Fatalf("separate sequence started at %d, want 1", n)
	}
}

func TestInMemorySequencerConcurrentGapFree(t *testing.T) {
	s := NewInMemorySequencer()
	const n = 200
	var wg sync.WaitGroup
	seen := make([]int64, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			v, _ := s.Next(context.Background(), "seq")
			seen[idx] = v
		}(i)
	}
	wg.Wait()

	counts := make(map[int64]int)
	for _, v := range seen {
		counts[v]++
	}
	for i := int64(1); i <= n; i++ {
		if counts[i] != 1 {
			t.Fatalf("value %d issued %d times (want exactly 1) — sequence not gap-free/unique", i, counts[i])
		}
	}
}
