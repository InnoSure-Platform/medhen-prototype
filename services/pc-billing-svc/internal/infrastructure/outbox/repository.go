package outbox

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/medhen/pc-billing-svc/internal/infrastructure/postgres"
)

type OutboxEvent struct {
	AggregateType string
	AggregateID   uuid.UUID
	EventType     string
	Payload       interface{}
}

type PostgresOutboxRepository struct{}

func NewPostgresOutboxRepository() *PostgresOutboxRepository {
	return &PostgresOutboxRepository{}
}

func (r *PostgresOutboxRepository) SaveEvent(ctx context.Context, event OutboxEvent) error {
	tx := postgres.ExtractTx(ctx)
	if tx == nil {
		return errors.New("transaction required for outbox save")
	}

	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), event.AggregateType, event.AggregateID, event.EventType, payloadBytes)

	return err
}
