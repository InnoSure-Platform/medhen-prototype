package domain

import "time"

// TopicPartyRegistered is the event/outbox topic for a newly registered party.
const TopicPartyRegistered = "party.registered"

// PartyRegistered is emitted when a party is registered. It is written to the
// outbox in the same transaction as the party row and later published to the
// event bus. It implements eventbus.Event via EventName.
type PartyRegistered struct {
	PartyID    string    `json:"party_id"`
	TenantID   string    `json:"tenant_id"`
	Type       Type      `json:"type"`
	FullName   string    `json:"full_name"`
	OccurredAt time.Time `json:"occurred_at"`
}

// EventName satisfies eventbus.Event.
func (PartyRegistered) EventName() string { return TopicPartyRegistered }
