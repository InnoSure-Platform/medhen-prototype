// Package kafka provides a minimal Kafka producer for Medhen domain events (ADR-PC-005).
package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// Publisher publishes JSON events to Kafka topics.
type Publisher struct {
	brokers []string
	writers map[string]*kafka.Writer
}

func NewPublisher(brokers ...string) *Publisher {
	if len(brokers) == 0 {
		brokers = []string{"localhost:19092"}
	}
	return &Publisher{brokers: brokers, writers: map[string]*kafka.Writer{}}
}

// NewPublisherFromEnv returns nil when KAFKA_BROKERS is unset (events disabled).
func NewPublisherFromEnv() *Publisher {
	raw := os.Getenv("KAFKA_BROKERS")
	if raw == "" {
		return nil
	}
	var brokers []string
	for _, b := range strings.Split(raw, ",") {
		if t := strings.TrimSpace(b); t != "" {
			brokers = append(brokers, t)
		}
	}
	if len(brokers) == 0 {
		return nil
	}
	return NewPublisher(brokers...)
}

func (p *Publisher) writer(topic string) *kafka.Writer {
	if w, ok := p.writers[topic]; ok {
		return w
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
	}
	p.writers[topic] = w
	return w
}

// Publish sends a JSON payload to the given topic.
func (p *Publisher) Publish(ctx context.Context, topic, key string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return p.writer(topic).WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: b,
		Time:  time.Now().UTC(),
	})
}

func (p *Publisher) Close() error {
	var err error
	for _, w := range p.writers {
		if e := w.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}

// RelayOutbox polls unpublished outbox rows and publishes them (at-least-once).
type OutboxRow struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       json.RawMessage
}

type OutboxFetcher func(ctx context.Context, limit int) ([]OutboxRow, error)
type OutboxMarker func(ctx context.Context, id string) error

func RelayOutbox(ctx context.Context, pub *Publisher, fetch OutboxFetcher, mark OutboxMarker, every time.Duration) {
	if pub == nil {
		slog.Info("kafka relay disabled — KAFKA_BROKERS not set")
		return
	}
	if every == 0 {
		every = 500 * time.Millisecond
	}
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			rows, err := fetch(ctx, 100)
			if err != nil {
				slog.Warn("outbox fetch", "err", err)
				continue
			}
			for _, r := range rows {
				topic := r.EventType
				var payload any
				_ = json.Unmarshal(r.Payload, &payload)
				if err := pub.Publish(ctx, topic, r.AggregateID, payload); err != nil {
					slog.Warn("kafka publish", "topic", topic, "err", err)
					continue
				}
				_ = mark(ctx, r.ID)
			}
		}
	}
}
