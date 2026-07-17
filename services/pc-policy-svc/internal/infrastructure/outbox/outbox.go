package outbox

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Event struct {
	ID            uuid.UUID
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte
}

type Repository interface {
	SaveWithTx(ctx context.Context, tx pgx.Tx, event *Event) error
}

type PostgresOutboxRepository struct{}

func NewPostgresOutboxRepository() *PostgresOutboxRepository {
	return &PostgresOutboxRepository{}
}

func (r *PostgresOutboxRepository) SaveWithTx(ctx context.Context, tx pgx.Tx, event *Event) error {
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(ctx, query, event.ID, event.AggregateType, event.AggregateID, event.EventType, payloadJSON)
	return err
}
