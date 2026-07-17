package notification

import (
	"time"
	"github.com/google/uuid"
)

type NotificationDispatchedEvent struct {
	NotificationID uuid.UUID `json:"notification_id"`
	PartyID        uuid.UUID `json:"party_id"`
	Channel        string    `json:"channel"`
	DispatchedAt   time.Time `json:"dispatched_at"`
}

type NotificationDeliveredEvent struct {
	NotificationID uuid.UUID `json:"notification_id"`
	ReceiptID      string    `json:"receipt_id"`
	DeliveredAt    time.Time `json:"delivered_at"`
}

type NotificationFailedEvent struct {
	NotificationID uuid.UUID `json:"notification_id"`
	Reason         string    `json:"reason"`
	FailedAt       time.Time `json:"failed_at"`
}
