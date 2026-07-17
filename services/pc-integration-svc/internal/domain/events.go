package domain

import (
	"time"

	"github.com/google/uuid"
)

// PaymentSettledEvent is published when a payment is fully confirmed by the provider.
type PaymentSettledEvent struct {
	EventID               uuid.UUID `json:"event_id"`
	InternalReferenceID   uuid.UUID `json:"internal_reference_id"`
	Provider              string    `json:"provider"`
	ProviderTransactionID string    `json:"provider_transaction_id"`
	AmountSettled         float64   `json:"amount_settled"`
	Currency              string    `json:"currency"`
	SettledAt             time.Time `json:"settled_at"`
}

// PaymentFailedEvent is published when a payment callback indicates failure.
type PaymentFailedEvent struct {
	EventID             uuid.UUID `json:"event_id"`
	InternalReferenceID uuid.UUID `json:"internal_reference_id"`
	Provider            string    `json:"provider"`
	Reason              string    `json:"reason"`
	FailedAt            time.Time `json:"failed_at"`
}
