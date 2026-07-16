package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/medhen/pc-party-mgmt-svc/internal/domain"
)

type OutboxTxRepository struct {
	tx pgx.Tx
}

func (r *OutboxTxRepository) Publish(ctx context.Context, event domain.DomainEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event payload: %w", err)
	}

	// partition_key helps ensure ordering in Kafka if needed (e.g. PartyID)
	// topic can be inferred from event type
	query := `INSERT INTO outbox (topic, partition_key, payload, created_at) VALUES ($1, $2, $3, $4)`
	
	_, err = r.tx.Exec(ctx, query, string(event.EventType()), event.EventID().String(), payload, event.OccurredAt())
	if err != nil {
		return fmt.Errorf("failed to insert outbox record: %w", err)
	}
	return nil
}
