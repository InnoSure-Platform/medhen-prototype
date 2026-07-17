package outbox

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/InnoSure-Platform/pc-shared-go-svc/events"
)

// Message represents the payload to be published to Kafka.
type Message struct {
	Topic        string
	PartitionKey string
	Payload      interface{}
}

// Publish writes an event to the outbox table within the provided transaction.
func Publish(ctx context.Context, tx pgx.Tx, msg Message) error {
	env, err := events.NewEnvelope("system", msg.Topic, msg.Payload)
	if err != nil {
		return err
	}

	payloadBytes, err := json.Marshal(env)
	if err != nil {
		return err
	}

	// Assuming the existence of an `outbox` table as specified in the platform standard.
	query := `
		INSERT INTO outbox (topic, partition_key, payload, created_at)
		VALUES ($1, $2, $3, NOW())
	`

	_, err = tx.Exec(ctx, query, msg.Topic, msg.PartitionKey, payloadBytes)
	return err
}
