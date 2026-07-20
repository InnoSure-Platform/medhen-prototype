// Package app holds the audit use cases: recording an immutable trail entry for
// every domain event, and querying the trail.
package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
)

// Record is one immutable audit-trail entry.
type Record struct {
	ID         string          `json:"id"`
	Topic      string          `json:"topic"`
	TenantID   string          `json:"tenant_id"`
	Payload    json.RawMessage `json:"payload"`
	RecordedAt time.Time       `json:"recorded_at"`
}

// Repository is the append-only audit store.
type Repository interface {
	Append(ctx context.Context, r Record) error
	List(ctx context.Context, tenantID, topic string, limit int) ([]Record, error)
}

// Recorder writes audit entries.
type Recorder struct{ repo Repository }

// NewRecorder builds the recorder.
func NewRecorder(repo Repository) *Recorder { return &Recorder{repo: repo} }

// Record appends a trail entry for an event, extracting the tenant from the
// payload when present. The event payload is stored verbatim.
func (r *Recorder) Record(ctx context.Context, topic string, payload []byte) error {
	var meta struct {
		TenantID string `json:"tenant_id"`
	}
	_ = json.Unmarshal(payload, &meta)

	return r.repo.Append(ctx, Record{
		ID: ids.New(), Topic: topic, TenantID: meta.TenantID,
		Payload: append(json.RawMessage(nil), payload...), RecordedAt: time.Now().UTC(),
	})
}

// List returns recent trail entries, optionally filtered by topic.
func (r *Recorder) List(ctx context.Context, tenantID, topic string, limit int) ([]Record, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return r.repo.List(ctx, tenantID, topic, limit)
}
