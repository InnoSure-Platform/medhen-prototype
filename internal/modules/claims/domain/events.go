package domain

import "time"

// Topics published by the claims module.
const (
	TopicClaimFiled   = "claims.filed"
	TopicClaimSettled = "claims.settled"
)

// ClaimFiled is emitted when a FNOL is recorded.
type ClaimFiled struct {
	ClaimID      string    `json:"claim_id"`
	TenantID     string    `json:"tenant_id"`
	PolicyID     string    `json:"policy_id"`
	PartyID      string    `json:"party_id"`
	ReserveMinor int64     `json:"reserve_minor"`
	OccurredAt   time.Time `json:"occurred_at"`
}

// EventName satisfies eventbus.Event.
func (ClaimFiled) EventName() string { return TopicClaimFiled }

// ClaimSettled is emitted when a claim is fast-track settled.
type ClaimSettled struct {
	ClaimID     string    `json:"claim_id"`
	TenantID    string    `json:"tenant_id"`
	PolicyID    string    `json:"policy_id"`
	PartyID     string    `json:"party_id"`
	AmountMinor int64     `json:"amount_minor"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// EventName satisfies eventbus.Event.
func (ClaimSettled) EventName() string { return TopicClaimSettled }
