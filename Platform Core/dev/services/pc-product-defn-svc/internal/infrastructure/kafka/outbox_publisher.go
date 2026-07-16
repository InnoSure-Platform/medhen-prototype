package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/medhen/pc-product-defn-svc/internal/domain/product"
)

// OutboxPublisher handles appending events to the transactional outbox table
// during the Postgres Unit of Work. A separate background worker (Relay) 
// using franz-go will pick these up and publish them to Kafka.
type OutboxPublisher struct{}

func NewOutboxPublisher() *OutboxPublisher {
	return &OutboxPublisher{}
}

// Publish stores the event in the outbox table using the provided transaction.
func (o *OutboxPublisher) Publish(ctx context.Context, tx pgx.Tx, event *product.ProductLifecycleEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal outbox event: %w", err)
	}

	topic := "platform.product.lifecycle.v1"
	key := fmt.Sprintf("%s:%s", event.TenantID, event.ProductID)

	query := `
		INSERT INTO outbox (id, topic, key, payload)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.Exec(ctx, query, uuid.New(), topic, key, payload)
	return err
}
