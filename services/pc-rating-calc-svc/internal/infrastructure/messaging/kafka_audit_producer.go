package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"
	"medhen.com/platform/pc-rating-calc-svc/internal/domain/models"
)

// KafkaAuditProducer implements the AuditEventProducer via franz-go
type KafkaAuditProducer struct {
	client *kgo.Client
	topic  string
}

func NewKafkaAuditProducer(brokers string) (*KafkaAuditProducer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers),
		kgo.ProducerBatchCompression(kgo.NoCompression()), // For latency
		kgo.RequiredAcks(kgo.LeaderAck()),                // Prioritize latency over strict durability for audit
	)
	if err != nil {
		return nil, err
	}

	return &KafkaAuditProducer{
		client: cl,
		topic:  "pc.rating.calculated.v1",
	}, nil
}

// PublishRatingEvent sends the audit event asynchronously
func (k *KafkaAuditProducer) PublishRatingEvent(ctx context.Context, breakdown *models.PremiumBreakdown, req models.RatingRequest) error {
	// Simple JSON mapping for demonstration; in reality, this would use Avro schema registry via Apicurio
	payload, err := json.Marshal(breakdown)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	record := &kgo.Record{
		Topic: k.topic,
		Key:   []byte(req.TenantID + ":" + breakdown.CalculationID),
		Value: payload,
	}

	k.client.Produce(ctx, record, func(r *kgo.Record, err error) {
		if err != nil {
			slog.Error("failed to publish rating audit event", "error", err, "calculation_id", breakdown.CalculationID)
		} else {
			slog.Info("Rating audit event dispatched", "calculation_id", breakdown.CalculationID, "partition", r.Partition)
		}
	})

	return nil
}

func (k *KafkaAuditProducer) Close() {
	k.client.Close()
}
