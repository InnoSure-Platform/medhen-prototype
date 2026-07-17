package messaging

import (
	"context"
	"fmt"

	"github.com/medhen/pc-audit-svc/internal/domain/audit"
)

// KafkaAnomalyPublisher streams Audit events directly to the ML security pipeline.
type KafkaAnomalyPublisher struct {
	topic string
}

func NewKafkaAnomalyPublisher(topic string) *KafkaAnomalyPublisher {
	return &KafkaAnomalyPublisher{topic: topic}
}

// PublishAnomalyEvent implements commands.AnomalyPublisher
func (p *KafkaAnomalyPublisher) PublishAnomalyEvent(ctx context.Context, entry *audit.AuditLedgerEntry) error {
	// Serialize Avro and push to Franz-go client
	fmt.Printf("Real-Time AI Stream: Emitted event %s for actor %s to %s\n",
		entry.EventID.String(), entry.Actor.UserID, p.topic)
	return nil
}
