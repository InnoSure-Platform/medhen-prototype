// Package outbox implements the transactional outbox pattern (ADR-PC-008).
package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Record struct {
	ID            string          `json:"id"`
	AggregateType string          `json:"aggregateType"`
	AggregateID   string          `json:"aggregateId"`
	EventType     string          `json:"eventType"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"createdAt"`
	PublishedAt   *time.Time      `json:"publishedAt,omitempty"`
}

const DDL = `
CREATE TABLE IF NOT EXISTS outbox (
  id UUID PRIMARY KEY,
  aggregate_type TEXT NOT NULL,
  aggregate_id TEXT NOT NULL,
  event_type TEXT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS outbox_unpublished_idx ON outbox (created_at) WHERE published_at IS NULL;
`

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, DDL)
	return err
}

// Insert writes an outbox row inside an existing transaction.
func Insert(ctx context.Context, tx pgx.Tx, aggregateType, aggregateID, eventType string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox (id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1,$2,$3,$4,$5)`,
		uuid.NewString(), aggregateType, aggregateID, eventType, b)
	return err
}

type Publisher func(ctx context.Context, rec Record) error

// Relay polls unpublished rows and publishes them (at-least-once).
func Relay(ctx context.Context, pool *pgxpool.Pool, pub Publisher, every time.Duration) {
	if every == 0 {
		every = 500 * time.Millisecond
	}
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_ = flush(ctx, pool, pub)
		}
	}
}

func flush(ctx context.Context, pool *pgxpool.Pool, pub Publisher) error {
	rows, err := pool.Query(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, created_at
		FROM outbox WHERE published_at IS NULL
		ORDER BY created_at LIMIT 100`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.ID, &r.AggregateType, &r.AggregateID, &r.EventType, &r.Payload, &r.CreatedAt); err != nil {
			return err
		}
		if err := pub(ctx, r); err != nil {
			continue
		}
		_, _ = pool.Exec(ctx, `UPDATE outbox SET published_at = now() WHERE id = $1`, r.ID)
	}
	return rows.Err()
}
