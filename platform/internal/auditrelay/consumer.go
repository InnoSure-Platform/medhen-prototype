package auditrelay

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/InnoSure-Platform/pc-platform/internal/store"
	"github.com/segmentio/kafka-go"
)

// ConsumeDomainEvents appends audit entries from Kafka topics (pc-audit-svc sidecar).
func ConsumeDomainEvents(ctx context.Context, brokers []string, repo store.Repository) {
	if len(brokers) == 0 {
		return
	}
	topics := []string{
		"pc.party.registered.v1",
		"pc.policy.quoted.v1",
		"pc.policy.bound.v1",
		"pc.policy.issued.v1",
		"pc.billing.payment.completed.v1",
		"pc.claim.registered.v1",
		"pc.claim.settled.v1",
	}
	for _, topic := range topics {
		go consume(ctx, brokers, topic, repo)
	}
}

func consume(ctx context.Context, brokers []string, topic string, repo store.Repository) {
	r := kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, Topic: topic, GroupID: "pc-audit-svc", MinBytes: 1, MaxBytes: 1e6})
	defer r.Close()
	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Warn("kafka read", "topic", topic, "err", err)
			time.Sleep(time.Second)
			continue
		}
		var payload map[string]any
		_ = json.Unmarshal(msg.Value, &payload)
		entityID, _ := payload["policyId"].(string)
		if entityID == "" {
			entityID, _ = payload["partyId"].(string)
		}
		if entityID == "" {
			entityID, _ = payload["quoteId"].(string)
		}
		if entityID == "" {
			entityID, _ = payload["claimId"].(string)
		}
		if entityID == "" {
			entityID = string(msg.Key)
		}
		_ = repo.AppendAudit(ctx, store.NewAuditEntry(topic, entityID, "EVENT", "kafka", string(msg.Value)))
	}
}
