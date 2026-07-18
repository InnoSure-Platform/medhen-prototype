// Package outbox implements the transactional outbox pattern. A command handler
// writes domain events to the outbox table in the SAME transaction as its
// aggregate changes (via the ambient Unit-of-Work), guaranteeing the event is
// persisted iff the state change commits. A Relay then drains unprocessed rows
// and hands them to a Publisher (the in-process event bus, or Kafka).
//
// The relay claims rows with `SELECT ... FOR UPDATE SKIP LOCKED` INSIDE the
// transaction that also marks them processed, so the row locks are held until
// commit. This makes concurrent relay workers safe (no double publishing) —
// unlike the pre-refactor relay whose lock was released before publishing.
package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
)

// Schema is the DDL for the outbox table.
const Schema = `
CREATE TABLE IF NOT EXISTS outbox (
    id             TEXT PRIMARY KEY,
    topic          TEXT NOT NULL,
    aggregate_type TEXT NOT NULL,
    aggregate_id   TEXT NOT NULL,
    payload        BYTEA NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed      BOOLEAN NOT NULL DEFAULT false,
    processed_at   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_outbox_unprocessed ON outbox (created_at) WHERE processed = false;
`

// Message is a single outbox entry.
type Message struct {
	ID            string
	Topic         string
	AggregateType string
	AggregateID   string
	Payload       []byte
}

// Write inserts a message using the ambient connection (the caller's tx). Call
// it inside the same database.WithinTx as the aggregate write.
func Write(ctx context.Context, q database.Querier, m Message) error {
	_, err := q.Exec(ctx,
		`INSERT INTO outbox (id, topic, aggregate_type, aggregate_id, payload)
		 VALUES ($1, $2, $3, $4, $5)`,
		m.ID, m.Topic, m.AggregateType, m.AggregateID, m.Payload)
	if err != nil {
		return fmt.Errorf("outbox: write: %w", err)
	}
	return nil
}

// Publisher delivers a message to its downstream (event bus / Kafka).
type Publisher interface {
	Publish(ctx context.Context, m Message) error
}

// PublisherFunc adapts a function to Publisher.
type PublisherFunc func(ctx context.Context, m Message) error

func (f PublisherFunc) Publish(ctx context.Context, m Message) error { return f(ctx, m) }

// Relay drains the outbox and publishes messages.
type Relay struct {
	db        *database.DB
	publisher Publisher
	batchSize int
	logger    *slog.Logger
}

// NewRelay builds a relay. batchSize<=0 defaults to 100.
func NewRelay(db *database.DB, publisher Publisher, batchSize int, logger *slog.Logger) *Relay {
	if batchSize <= 0 {
		batchSize = 100
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Relay{db: db, publisher: publisher, batchSize: batchSize, logger: logger}
}

// ProcessBatch claims up to batchSize unprocessed messages, publishes them, and
// marks them processed — all in one transaction. Returns the number processed.
func (r *Relay) ProcessBatch(ctx context.Context) (int, error) {
	var processed int
	err := r.db.WithinTx(ctx, func(ctx context.Context) error {
		q := r.db.Conn(ctx)

		rows, err := q.Query(ctx,
			`SELECT id, topic, aggregate_type, aggregate_id, payload
			   FROM outbox
			  WHERE processed = false
			  ORDER BY created_at
			  FOR UPDATE SKIP LOCKED
			  LIMIT $1`, r.batchSize)
		if err != nil {
			return fmt.Errorf("outbox: select: %w", err)
		}

		// Fully drain rows before issuing further statements on this tx conn.
		var batch []Message
		for rows.Next() {
			var m Message
			if err := rows.Scan(&m.ID, &m.Topic, &m.AggregateType, &m.AggregateID, &m.Payload); err != nil {
				rows.Close()
				return fmt.Errorf("outbox: scan: %w", err)
			}
			batch = append(batch, m)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("outbox: rows: %w", err)
		}

		for _, m := range batch {
			if err := r.publisher.Publish(ctx, m); err != nil {
				// Abort the tx: nothing is marked processed, locks release on
				// rollback, and the batch is retried on the next tick.
				return fmt.Errorf("outbox: publish %s: %w", m.ID, err)
			}
			if _, err := q.Exec(ctx,
				`UPDATE outbox SET processed = true, processed_at = now() WHERE id = $1`, m.ID); err != nil {
				return fmt.Errorf("outbox: mark %s: %w", m.ID, err)
			}
			processed++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return processed, nil
}

// Run polls ProcessBatch on the given interval until ctx is cancelled.
func (r *Relay) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n, err := r.ProcessBatch(ctx); err != nil {
				r.logger.Error("outbox relay batch failed", "err", err)
			} else if n > 0 {
				r.logger.Debug("outbox relay published", "count", n)
			}
		}
	}
}
