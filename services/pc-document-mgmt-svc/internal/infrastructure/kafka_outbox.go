package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/inno-sphere/medhen-prototype/services/pc-document-mgmt-svc/internal/domain"
	"database/sql"
)

type KafkaOutboxPublisher struct {
	db *sql.DB
}

func NewKafkaOutboxPublisher(db *sql.DB) *KafkaOutboxPublisher {
	return &KafkaOutboxPublisher{db: db}
}

func (p *KafkaOutboxPublisher) Publish(ctx context.Context, event domain.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	query := `
		INSERT INTO outbox (
			topic, partition_key, payload, headers
		) VALUES ($1, $2, $3, '{}')`

	_, err = p.db.ExecContext(ctx, query, event.Topic(), event.PartitionKey(), payload)
	if err != nil {
		return fmt.Errorf("failed to write outbox: %w", err)
	}

	return nil
}
