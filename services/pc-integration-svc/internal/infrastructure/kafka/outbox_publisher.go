package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/medhen/pc-integration-svc/internal/application/ports"
	"github.com/medhen/pc-integration-svc/internal/domain"
	"github.com/twmb/franz-go/pkg/kgo"
)

type OutboxPublisher struct {
	client *kgo.Client
}

func NewOutboxPublisher(client *kgo.Client) ports.EventPublisher {
	return &OutboxPublisher{client: client}
}

func (p *OutboxPublisher) PublishPaymentSettled(ctx context.Context, event *domain.PaymentSettledEvent) error {
	// In a real transactional outbox, this would be inserted into a postgres `outbox` table first.
	// For the MVP, we publish directly to Kafka to simulate the final delivery.
	
	payload, err := json.Marshal(event) // Should be Avro in production per the spec
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	record := &kgo.Record{
		Topic: "platform.integration.payment.settled.v1",
		Key:   []byte(event.InternalReferenceID.String()),
		Value: payload,
	}

	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("failed to produce to kafka: %w", err)
	}

	return nil
}

func (p *OutboxPublisher) PublishPaymentFailed(ctx context.Context, event *domain.PaymentFailedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	record := &kgo.Record{
		Topic: "platform.integration.payment.failed.v1",
		Key:   []byte(event.InternalReferenceID.String()),
		Value: payload,
	}

	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		return fmt.Errorf("failed to produce to kafka: %w", err)
	}

	return nil
}
