package domain

import "time"

// TopicPolicyIssued is the event/outbox topic for a newly issued policy.
const TopicPolicyIssued = "policy.issued"

// PolicyIssued is emitted (via the outbox, in the bind transaction) when a
// policy is issued. Billing consumes it to raise the first invoice.
type PolicyIssued struct {
	PolicyID     string    `json:"policy_id"`
	PolicyNumber string    `json:"policy_number"`
	TenantID     string    `json:"tenant_id"`
	PartyID      string    `json:"party_id"`
	ProductCode  string    `json:"product_code"`
	GrossMinor   int64     `json:"gross_minor"`
	OccurredAt   time.Time `json:"occurred_at"`
}

// EventName satisfies eventbus.Event.
func (PolicyIssued) EventName() string { return TopicPolicyIssued }
