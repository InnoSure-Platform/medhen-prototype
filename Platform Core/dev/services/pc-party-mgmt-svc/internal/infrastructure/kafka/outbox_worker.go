package kafka

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxWorker struct {
	db        *pgxpool.Pool
	publisher *OutboxPublisher
	interval  time.Duration
}

func NewOutboxWorker(db *pgxpool.Pool, publisher *OutboxPublisher, interval time.Duration) *OutboxWorker {
	return &OutboxWorker{
		db:        db,
		publisher: publisher,
		interval:  interval,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down outbox worker")
			return
		case <-ticker.C:
			w.processOutbox(ctx)
		}
	}
}

func (w *OutboxWorker) processOutbox(ctx context.Context) {
	// Query up to 50 unpublished events
	query := `
		SELECT id, topic, partition_key, payload 
		FROM outbox 
		WHERE published_at IS NULL 
		ORDER BY created_at ASC 
		LIMIT 50
	`

	rows, err := w.db.Query(ctx, query)
	if err != nil {
		slog.Error("Failed to query outbox", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var topic, partitionKey string
		var payload []byte

		if err := rows.Scan(&id, &topic, &partitionKey, &payload); err != nil {
			slog.Error("Failed to scan outbox row", "error", err)
			continue
		}

		// Publish to Kafka
		err := w.publisher.Publish(ctx, topic, partitionKey, payload)
		if err != nil {
			slog.Error("Failed to publish outbox event, will retry", "id", id, "error", err)
			// Break to respect ordering, retry on next tick
			break
		}

		// Mark as published
		updateQuery := `UPDATE outbox SET published_at = now() WHERE id = $1`
		_, err = w.db.Exec(ctx, updateQuery, id)
		if err != nil {
			slog.Error("Failed to mark outbox as published, event may be duplicated", "id", id, "error", err)
			// Breaking to avoid massive duplication
			break 
		}
	}
}
