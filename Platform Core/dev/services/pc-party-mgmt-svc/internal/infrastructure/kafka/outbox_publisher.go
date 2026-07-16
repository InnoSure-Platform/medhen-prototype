package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type OutboxPublisher struct {
	writer *kafka.Writer
}

// NewOutboxPublisher creates a real Kafka publisher. 
// In a full implementation, this integrates with Schema Registry for Avro.
func NewOutboxPublisher(brokers []string) *OutboxPublisher {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
		// Async: false, to guarantee it is written before we mark as published
	}

	return &OutboxPublisher{
		writer: w,
	}
}

// Publish actually pushes the event to the Kafka broker.
func (o *OutboxPublisher) Publish(ctx context.Context, topic, partitionKey string, payload []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(partitionKey),
		Value: payload,
	}

	err := o.writer.WriteMessages(ctx, msg)
	if err != nil {
		slog.Error("Failed to write message to Kafka", "topic", topic, "error", err)
		return fmt.Errorf("failed to publish to kafka: %w", err)
	}

	slog.Info("Successfully published event to Kafka", "topic", topic, "key", partitionKey)
	return nil
}

func (o *OutboxPublisher) Close() error {
	return o.writer.Close()
}
