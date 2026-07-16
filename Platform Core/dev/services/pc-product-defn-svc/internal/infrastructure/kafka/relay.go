package kafka

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sony/gobreaker"
	"github.com/twmb/franz-go/pkg/kgo"
)

// OutboxRelay polls the outbox table and publishes events to Kafka.
type OutboxRelay struct {
	db      *pgxpool.Pool
	client  *kgo.Client
	breaker *gobreaker.CircuitBreaker
}

// NewOutboxRelay creates a new OutboxRelay.
func NewOutboxRelay(db *pgxpool.Pool, brokers []string) (*OutboxRelay, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ProducerBatchCompression(kgo.ZstdCompression()),
	)
	if err != nil {
		return nil, err
	}

	// Circuit breaker configuration for Kafka publishing
	cbSettings := gobreaker.Settings{
		Name:        "KafkaPublish",
		MaxRequests: 5,
		Interval:    10 * time.Second,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.6
		},
	}

	return &OutboxRelay{
		db:      db,
		client:  client,
		breaker: gobreaker.NewCircuitBreaker(cbSettings),
	}, nil
}

// Start runs the polling loop in the background.
func (r *OutboxRelay) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.client.Close()
			return
		case <-ticker.C:
			if err := r.processOutbox(ctx); err != nil {
				slog.Error("Failed to process outbox", "error", err)
			}
		}
	}
}

func (r *OutboxRelay) processOutbox(ctx context.Context) error {
	// 1. Fetch unpublished events
	query := `
		SELECT id, topic, key, payload 
		FROM outbox 
		ORDER BY created_at ASC 
		LIMIT 50
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var publishedIDs []string

	for rows.Next() {
		var id, topic, key string
		var payload []byte
		if err := rows.Scan(&id, &topic, &key, &payload); err != nil {
			slog.Error("Failed to scan outbox row", "error", err)
			continue
		}

		// 2. Publish to Kafka wrapped in Circuit Breaker
		_, err := r.breaker.Execute(func() (interface{}, error) {
			record := &kgo.Record{
				Topic: topic,
				Key:   []byte(key),
				Value: payload,
			}
			
			// Synchronous publish for strong guarantee in this worker
			if err := r.client.ProduceSync(ctx, record).FirstErr(); err != nil {
				return nil, err
			}
			return nil, nil
		})

		if err != nil {
			slog.Error("Circuit breaker tripped or publish failed", "error", err)
			break // Stop processing this batch
		}

		publishedIDs = append(publishedIDs, id)
	}

	// 3. Delete published events from outbox
	if len(publishedIDs) > 0 {
		_, err := r.db.Exec(ctx, "DELETE FROM outbox WHERE id = ANY($1)", publishedIDs)
		if err != nil {
			slog.Error("Failed to delete published outbox events", "error", err)
		}
	}

	return nil
}
