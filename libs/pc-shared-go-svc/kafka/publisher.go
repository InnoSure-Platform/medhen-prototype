package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

// Publisher defines an interface for publishing messages to Kafka.
type Publisher interface {
	Publish(ctx context.Context, topic, key string, payload interface{}) error
}

// WriterPublisher implements Publisher using kafka-go.
type WriterPublisher struct {
	writer *kafka.Writer
}

// NewPublisher creates a new Kafka publisher.
func NewPublisher(brokers []string) *WriterPublisher {
	return &WriterPublisher{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			AllowAutoTopicCreation: true,
			Balancer:               &kafka.Hash{}, // Guarantees ordering per partition key
		},
	}
}

// Publish publishes a message, automatically injecting OpenTelemetry headers.
func (p *WriterPublisher) Publish(ctx context.Context, topic, key string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	}

	// Inject OpenTelemetry headers
	otel.GetTextMapPropagator().Inject(ctx, NewMessageCarrier(&msg))

	return p.writer.WriteMessages(ctx, msg)
}

// Close closes the underlying writer.
func (p *WriterPublisher) Close() error {
	return p.writer.Close()
}

// MessageCarrier adapts a kafka.Message to propagation.TextMapCarrier.
type MessageCarrier struct {
	msg *kafka.Message
}

func NewMessageCarrier(msg *kafka.Message) *MessageCarrier {
	return &MessageCarrier{msg: msg}
}

func (c *MessageCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *MessageCarrier) Set(key string, value string) {
	c.msg.Headers = append(c.msg.Headers, kafka.Header{
		Key:   key,
		Value: []byte(value),
	})
}

func (c *MessageCarrier) Keys() []string {
	keys := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		keys[i] = h.Key
	}
	return keys
}
