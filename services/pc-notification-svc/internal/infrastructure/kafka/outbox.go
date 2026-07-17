package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

type OutboxPublisher struct {
	client *kgo.Client
}

func NewOutboxPublisher(brokers []string) (*OutboxPublisher, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ProducerBatchCompression(kgo.ZstdCompression()),
	)
	if err != nil {
		return nil, err
	}
	return &OutboxPublisher{client: client}, nil
}

func (p *OutboxPublisher) PublishEvent(ctx context.Context, topic string, key string, event interface{}) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	record := &kgo.Record{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
	}

	// Synchronous publish (could be async in a real high-throughput scenario, or transactionally tied to PG Outbox)
	results := p.client.ProduceSync(ctx, record)
	if results.FirstErr() != nil {
		return fmt.Errorf("kafka produce failed: %w", results.FirstErr())
	}

	return nil
}

func (p *OutboxPublisher) Close() {
	p.client.Close()
}
