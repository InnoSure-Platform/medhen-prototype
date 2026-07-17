package kafka

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type OutboxRelay struct {
	db     *sql.DB
	writer *kafka.Writer
	logger *slog.Logger
}

func NewOutboxRelay(db *sql.DB, brokers []string, logger *slog.Logger) *OutboxRelay {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.LeastBytes{},
	}
	return &OutboxRelay{
		db:     db,
		writer: w,
		logger: logger,
	}
}

func (r *OutboxRelay) Start(ctx context.Context, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	defer r.writer.Close()

	r.logger.Info("Starting Kafka Outbox Relay worker")

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Stopping Outbox Relay worker")
			return
		case <-ticker.C:
			r.processOutbox(ctx)
		}
	}
}

func (r *OutboxRelay) processOutbox(ctx context.Context) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction for outbox", "error", err)
		return
	}
	defer tx.Rollback()

	// Select 100 rows, locking them to prevent concurrent relay workers from publishing duplicates
	query := `
		SELECT id, topic, partition_key, payload 
		FROM outbox 
		ORDER BY created_at ASC 
		LIMIT 100 
		FOR UPDATE SKIP LOCKED`
	
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to query outbox", "error", err)
		return
	}
	defer rows.Close()

	var messages []kafka.Message
	var ids []int

	for rows.Next() {
		var id int
		var topic, key string
		var payload []byte
		
		if err := rows.Scan(&id, &topic, &key, &payload); err != nil {
			r.logger.Error("Failed to scan outbox row", "error", err)
			continue
		}
		
		messages = append(messages, kafka.Message{
			Topic: topic,
			Key:   []byte(key),
			Value: payload,
		})
		ids = append(ids, id)
	}

	if len(messages) == 0 {
		return // Nothing to process
	}

	// Publish to Kafka
	if err := r.writer.WriteMessages(ctx, messages...); err != nil {
		r.logger.Error("Failed to publish messages to Kafka", "error", err)
		return
	}

	// Delete published rows
	for _, id := range ids {
		if _, err := tx.ExecContext(ctx, "DELETE FROM outbox WHERE id = $1", id); err != nil {
			r.logger.Error("Failed to delete outbox row", "id", id, "error", err)
			return // Trigger rollback
		}
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit outbox transaction", "error", err)
	} else {
		r.logger.Info("Relayed events to Kafka", "count", len(messages))
	}
}
