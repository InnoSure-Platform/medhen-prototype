package store

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
)

// FetchOutbox returns unpublished rows for the Kafka relay.
func (r *PostgresRepository) FetchOutbox(ctx context.Context, limit int) ([]OutboxRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload
		FROM outbox WHERE published_at IS NULL ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []OutboxRow
	for rows.Next() {
		var row OutboxRow
		if err := rows.Scan(&row.ID, &row.AggregateType, &row.AggregateID, &row.EventType, &row.Payload); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) MarkOutboxPublished(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE outbox SET published_at = now() WHERE id = $1`, id)
	return err
}

// OutboxRow is a pending domain event.
type OutboxRow struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       json.RawMessage
}

// Bridge kafka package without import cycle — used by runtime.
type OutboxFetcher = func(ctx context.Context, limit int) ([]OutboxRow, error)

// Ensure pgx import used
var _ = pgx.ErrNoRows
