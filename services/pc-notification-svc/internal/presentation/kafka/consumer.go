package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"

	"pc-notification-svc/internal/application/command"
)

type Consumer struct {
	logger  *slog.Logger
	handler *command.DispatchHandler
	client  *kgo.Client
}

func NewConsumer(logger *slog.Logger, brokers []string, groupID string, topics []string, h *command.DispatchHandler) (*Consumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topics...),
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		logger:  logger,
		handler: h,
		client:  cl,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	c.logger.Info("Starting Kafka consumer group...")

	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return
		}
		
		fetches.EachError(func(topic string, partition int32, err error) {
			c.logger.Error("Kafka fetch error", "topic", topic, "partition", partition, "error", err)
		})

		fetches.EachRecord(func(record *kgo.Record) {
			c.logger.Debug("Received record", "topic", record.Topic)
			c.HandleMessage(ctx, record.Topic, record.Value)
		})
	}
}

func (c *Consumer) HandleMessage(ctx context.Context, topic string, payload []byte) {
	var event map[string]interface{}
	if err := json.Unmarshal(payload, &event); err != nil {
		c.logger.Error("Failed to unmarshal event", "error", err)
		return
	}

	eventName := "UnknownEvent"
	if topic == "platform.policy.lifecycle.v1" {
		eventName = "PolicyBound" // Simplified
	} else if topic == "platform.party.lifecycle.v1" {
		eventName = "PartyErased"
	}

	partyIDStr, _ := event["party_id"].(string)
	partyID, _ := uuid.Parse(partyIDStr)

	// Dispatch Command
	cmd := command.DispatchCommand{
		TenantID:     "t-default",
		PartyID:      partyID,
		EventName:    eventName,
		Payload:      event,
		TargetLocale: "en-US",
	}

	c.logger.Info("Consumed event, dispatching notification", "event", eventName)
	if err := c.handler.Handle(ctx, cmd); err != nil {
		c.logger.Error("Dispatch handler failed", "error", err)
	}
}

func (c *Consumer) Close() {
	c.client.Close()
}
