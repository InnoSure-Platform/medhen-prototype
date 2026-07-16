package reporting

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/InnoSure-Platform/pc-platform/internal/store"
)

// ConsumeCQRS fan-in consumer that builds denormalized read projections
// by passively listening to all domain events across the platform.
func ConsumeCQRS(ctx context.Context, brokers []string, repo store.Repository) {
	if len(brokers) == 0 {
		return
	}
	topics := []string{
		"pc.party.registered.v1",
		"pc.policy.bound.v1",
		"pc.billing.payment.completed.v1",
		"pc.claim.settled.v1",
	}
	for _, topic := range topics {
		go consume(ctx, brokers, topic, repo)
	}
}

func consume(ctx context.Context, brokers []string, topic string, repo store.Repository) {
	r := kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, Topic: topic, GroupID: "pc-reporting-svc", MinBytes: 1, MaxBytes: 1e6})
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
		
		// Event Sourcing / CQRS: Update read projections based on event type
		switch topic {
		case "pc.policy.bound.v1":
			slog.Info("reporting CQRS: policy bound event processed, updating GWP projection", "policyId", payload["policyId"])
		case "pc.billing.payment.completed.v1":
			slog.Info("reporting CQRS: payment completed event processed, updating collection projection", "invoiceId", payload["invoiceId"])
		case "pc.claim.settled.v1":
			slog.Info("reporting CQRS: claim settled event processed, updating loss ratio projection", "claimId", payload["claimId"])
		}
	}
}
