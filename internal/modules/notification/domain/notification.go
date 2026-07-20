// Package domain is the notification bounded context: outbound messages queued in
// response to domain events and dispatched to a provider.
package domain

import (
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
)

// Channel is the delivery channel.
type Channel string

const (
	ChannelSMS   Channel = "SMS"
	ChannelEmail Channel = "EMAIL"
)

// Status is the delivery state.
type Status string

const (
	StatusQueued Status = "QUEUED"
	StatusSent   Status = "SENT"
	StatusFailed Status = "FAILED"
)

// Notification is an outbound message.
type Notification struct {
	ID        string
	TenantID  string
	Channel   Channel
	Recipient string
	Subject   string
	Body      string
	Status    Status
	Attempts  int
	CreatedAt time.Time
	SentAt    *time.Time
}

// NewSMS queues an SMS notification.
func NewSMS(tenantID, recipient, body string) *Notification {
	return &Notification{
		ID: ids.New(), TenantID: tenantID, Channel: ChannelSMS, Recipient: recipient,
		Body: body, Status: StatusQueued, CreatedAt: time.Now().UTC(),
	}
}

// MarkSent records a successful delivery.
func (n *Notification) MarkSent() {
	now := time.Now().UTC()
	n.Status = StatusSent
	n.SentAt = &now
	n.Attempts++
}

// MarkFailed records a failed attempt.
func (n *Notification) MarkFailed() {
	n.Status = StatusFailed
	n.Attempts++
}
