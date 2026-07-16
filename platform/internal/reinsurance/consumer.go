package reinsurance

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/InnoSure-Platform/pc-platform/internal/store"
)

// ConsumeDomainEvents subscribes to events that impact Reinsurance cessions.
func ConsumeDomainEvents(ctx context.Context, brokers []string, repo store.Repository) {
	if len(brokers) == 0 {
		return
	}
	go consume(ctx, brokers, "pc.policy.bound.v1", repo)
	go consume(ctx, brokers, "pc.claim.settled.v1", repo)
}

func consume(ctx context.Context, brokers []string, topic string, repo store.Repository) {
	r := kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, Topic: topic, GroupID: "pc-reinsurance-svc", MinBytes: 1, MaxBytes: 1e6})
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
		
		if topic == "pc.policy.bound.v1" {
			policyID, _ := payload["policyId"].(string)
			slog.Info("reinsurance: computed automatic cession for policy", "policyId", policyID)
			// Here we would lookup the applicable Treaty and calculate the ceded premium
			// e.g. record cession in pc_reinsurance schema
		} else if topic == "pc.claim.settled.v1" {
			claimID, _ := payload["claimId"].(string)
			slog.Info("reinsurance: computed recovery for claim", "claimId", claimID)
			// Calculate reinsurance recovery amount for settled claims
		}
	}
}
