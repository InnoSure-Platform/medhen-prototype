package outbox

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// KafkaProducer interface allows injecting an actual Kafka client
type KafkaProducer interface {
	Publish(ctx context.Context, topic, key string, payload []byte) error
}

type RelayWorker struct {
	pool     *pgxpool.Pool
	producer KafkaProducer
	interval time.Duration
}

func NewRelayWorker(pool *pgxpool.Pool, producer KafkaProducer) *RelayWorker {
	return &RelayWorker{
		pool:     pool,
		producer: producer,
		interval: 1 * time.Second, // Default polling interval
	}
}

// Start begins the continuous polling loop
func (r *RelayWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Outbox relay shutting down...")
			return
		case <-ticker.C:
			r.processOutbox(ctx)
		}
	}
}

func (r *RelayWorker) processOutbox(ctx context.Context) {
	// 1. Fetch unpublished events, locking them for update to prevent concurrent worker clashes
	rows, err := r.pool.Query(ctx, `
		SELECT id, topic, partition_key, payload 
		FROM outbox 
		WHERE processed = FALSE 
		ORDER BY id ASC 
		LIMIT 100 
		FOR UPDATE SKIP LOCKED
	`)
	
	if err != nil {
		log.Printf("Outbox relay query error: %v", err)
		return
	}
	defer rows.Close()

	type OutboxEvent struct {
		ID           int64
		Topic        string
		PartitionKey string
		Payload      []byte
	}
	var events []OutboxEvent

	for rows.Next() {
		var evt OutboxEvent
		if err := rows.Scan(&evt.ID, &evt.Topic, &evt.PartitionKey, &evt.Payload); err != nil {
			log.Printf("Outbox row scan error: %v", err)
			continue
		}
		events = append(events, evt)
	}
	rows.Close() // Ensure connection is released before the update loop

	if len(events) == 0 {
		return // Nothing to process
	}

	// 2. Publish to Kafka and mark as processed
	for _, evt := range events {
		err := r.producer.Publish(ctx, evt.Topic, evt.PartitionKey, evt.Payload)
		if err != nil {
			// Backoff and retry later; don't mark as processed
			log.Printf("Failed to publish outbox event ID %d: %v", evt.ID, err)
			continue 
		}

		// 3. Mark processed
		_, err = r.pool.Exec(ctx, `UPDATE outbox SET processed = TRUE WHERE id = $1`, evt.ID)
		if err != nil {
			log.Printf("Failed to mark outbox event ID %d as processed: %v", evt.ID, err)
		}
	}
}
