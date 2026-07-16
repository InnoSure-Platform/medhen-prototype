package policyrelay

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/InnoSure-Platform/pc-platform/internal/usecase"
)

// ConsumeDomainEvents subscribes to events that impact the Policy boundary.
func ConsumeDomainEvents(ctx context.Context, brokers []string, m *usecase.Motor) {
	if len(brokers) == 0 {
		return
	}
	// policy-svc needs to know when a claim is settled
	go consume(ctx, brokers, "pc.claim.settled.v1", m)
}

func consume(ctx context.Context, brokers []string, topic string, m *usecase.Motor) {
	r := kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, Topic: topic, GroupID: "pc-policy-svc", MinBytes: 1, MaxBytes: 1e6})
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
		
		claimID, _ := payload["claimId"].(string)
		if claimID == "" {
			continue
		}

		// When a claim is settled, check if it's TOTAL_LOSS and cancel the policy
		claim, err := m.Repo.GetClaim(ctx, claimID)
		if err != nil {
			slog.Error("failed to fetch claim for event", "claimId", claimID, "err", err)
			continue
		}
		
		if claim.Track == "TOTAL_LOSS" {
			_ = m.CancelPolicy(ctx, usecase.CancelPolicyCmd{
				PolicyID: claim.PolicyID,
				Actor:    "system", // automatic cancellation by system via EDA
				Reason:   "Total Loss Exhaustion (EDA)",
			})
			slog.Info("policy cancelled automatically via event", "policyId", claim.PolicyID, "claimId", claim.ID)
		}
	}
}
