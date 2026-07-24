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

// Servicing event topics.
const (
	TopicPolicyEndorsed  = "policy.endorsed"
	TopicPolicyCancelled = "policy.cancelled"
	TopicPolicyRenewed   = "policy.renewed"
)

// PolicyEndorsed is emitted when an in-force policy's premium is adjusted.
type PolicyEndorsed struct {
	PolicyID      string    `json:"policy_id"`
	TenantID      string    `json:"tenant_id"`
	DeltaMinor    int64     `json:"delta_minor"`
	NewGrossMinor int64     `json:"new_gross_minor"`
	Reason        string    `json:"reason"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// EventName satisfies eventbus.Event.
func (PolicyEndorsed) EventName() string { return TopicPolicyEndorsed }

// PolicyCancelled is emitted when a policy is cancelled.
type PolicyCancelled struct {
	PolicyID    string    `json:"policy_id"`
	TenantID    string    `json:"tenant_id"`
	Reason      string    `json:"reason"`
	RefundMinor int64     `json:"refund_minor"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// EventName satisfies eventbus.Event.
func (PolicyCancelled) EventName() string { return TopicPolicyCancelled }

// PolicyRenewed is emitted when a successor policy is created from a prior one.
type PolicyRenewed struct {
	PolicyID      string    `json:"policy_id"`
	PriorPolicyID string    `json:"prior_policy_id"`
	PolicyNumber  string    `json:"policy_number"`
	TenantID      string    `json:"tenant_id"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// EventName satisfies eventbus.Event.
func (PolicyRenewed) EventName() string { return TopicPolicyRenewed }
