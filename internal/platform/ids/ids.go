// Package ids provides identifier generation for the platform: sortable ULIDs
// for entity keys, and a Sequencer abstraction for monotonic business numbers
// such as policy numbers (EIC/MOT/{year}/{seq}).
package ids

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"

	"github.com/oklog/ulid/v2"
)

// New returns a new lexicographically-sortable ULID string.
func New() string {
	return ulid.Make().String()
}

// NewWithPrefix returns a prefixed ULID, e.g. NewWithPrefix("pol") → "pol_01H...".
func NewWithPrefix(prefix string) string {
	return prefix + "_" + New()
}

// Sequencer issues gap-free monotonic numbers per named sequence (e.g. per LOB
// and year). Production uses a Postgres-backed implementation (added with the
// database module); InMemorySequencer is for tests and single-process dev.
type Sequencer interface {
	Next(ctx context.Context, name string) (int64, error)
}

// InMemorySequencer is a concurrency-safe in-process Sequencer. It does NOT
// survive restarts and must not be used across replicas.
type InMemorySequencer struct {
	mu       sync.Mutex
	counters map[string]int64
}

// NewInMemorySequencer creates an empty in-memory sequencer.
func NewInMemorySequencer() *InMemorySequencer {
	return &InMemorySequencer{counters: make(map[string]int64)}
}

// Next returns the next value for the named sequence, starting at 1.
func (s *InMemorySequencer) Next(_ context.Context, name string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[name]++
	return s.counters[name], nil
}

// PolicyNumber formats an insurer policy number as INSURER/LOB/YEAR/SEQ with the
// sequence zero-padded to 6 digits, e.g. "EIC/MOT/2026/000042".
func PolicyNumber(insurer, lob string, year int, seq int64) string {
	return fmt.Sprintf("%s/%s/%d/%06d", insurer, lob, year, seq)
}

// entropy is retained to make the dependency explicit; ulid.Make uses a
// monotonic reader internally, but this documents our crypto/rand source.
var _ = rand.Reader
