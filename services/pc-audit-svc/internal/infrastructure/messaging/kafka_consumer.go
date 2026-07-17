package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/medhen/pc-audit-svc/internal/application/commands"
	"github.com/medhen/pc-audit-svc/internal/domain/audit"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaConsumer struct {
	client        *kgo.Client
	appendHandler *commands.AppendRecordHandler
}

func NewKafkaConsumer(brokers []string, appendHandler *commands.AppendRecordHandler) (*KafkaConsumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics("platform.party.lifecycle.v1", "platform.policy.cdc.v1"),
		kgo.ConsumerGroup("pc-audit-svc-cdc-consumer"),
	)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		client:        client,
		appendHandler: appendHandler,
	}, nil
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	go func() {
		for {
			fetches := c.client.PollFetches(ctx)
			if fetches.IsClientClosed() {
				return
			}
			fetches.EachError(func(t string, p int32, err error) {
				fmt.Printf("Kafka fetch error topic %s partition %d: %v\n", t, p, err)
			})

			fetches.EachRecord(func(record *kgo.Record) {
				// Parse Kafka record and execute command
				// In reality, this would decode Avro via Schema Registry

				var payload map[string]interface{}
				if err := json.Unmarshal(record.Value, &payload); err == nil {
					cmd := commands.AppendRecordCommand{
						TenantID:       "default-tenant", // Extract from headers/payload
						Actor:          audit.ActorContext{UserID: "system", Role: "cdc"},
						ActionType:     "CDC_UPDATE",
						EntityType:     string(record.Topic),
						EntityID:       string(record.Key),
						IsPIIEncrypted: false,
						DeltaPlaintext: record.Value,
					}

					if err := c.appendHandler.Handle(ctx, cmd); err != nil {
						fmt.Printf("Failed to append record from kafka: %v\n", err)
					}
				}
			})
		}
	}()
}

func (c *KafkaConsumer) Close() {
	c.client.Close()
}
