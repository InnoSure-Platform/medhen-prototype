package notification

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string
type Channel string
type Category string

const (
	StatusPending    Status = "PENDING"
	StatusDispatched Status = "DISPATCHED"
	StatusDelivered  Status = "DELIVERED"
	StatusFailed     Status = "FAILED"
	StatusSuppressed Status = "SUPPRESSED"

	ChannelSMS   Channel = "SMS"
	ChannelEmail Channel = "EMAIL"
	ChannelInApp Channel = "IN_APP"

	CategoryMarketing     Category = "MARKETING"
	CategoryTransactional Category = "TRANSACTIONAL"
	CategoryStatutory     Category = "STATUTORY"
)

type Notification struct {
	ID               uuid.UUID
	TenantID         string
	PartyID          uuid.UUID
	TemplateCode     string
	Channel          Channel
	Category         Category
	Status           Status
	RecipientAddress string
	RenderedContent  string
	VendorReceiptID  *string
	ErrorReason      *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewNotification creates a new PENDING notification
func NewNotification(tenantID string, partyID uuid.UUID, tplCode string, channel Channel, category Category, address, content string) *Notification {
	now := time.Now().UTC()
	return &Notification{
		ID:               uuid.New(),
		TenantID:         tenantID,
		PartyID:          partyID,
		TemplateCode:     tplCode,
		Channel:          channel,
		Category:         category,
		Status:           StatusPending,
		RecipientAddress: address,
		RenderedContent:  content,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func (n *Notification) MarkDispatched() error {
	if n.Status != StatusPending {
		return errors.New("only PENDING notifications can be dispatched")
	}
	n.Status = StatusDispatched
	n.UpdatedAt = time.Now().UTC()
	return nil
}

func (n *Notification) MarkDelivered(receiptID string) error {
	if n.Status != StatusDispatched {
		return errors.New("only DISPATCHED notifications can be marked DELIVERED")
	}
	n.Status = StatusDelivered
	n.VendorReceiptID = &receiptID
	n.UpdatedAt = time.Now().UTC()
	return nil
}

func (n *Notification) MarkFailed(reason string) {
	n.Status = StatusFailed
	n.ErrorReason = &reason
	n.UpdatedAt = time.Now().UTC()
}

func (n *Notification) MarkSuppressed(reason string) {
	n.Status = StatusSuppressed
	n.ErrorReason = &reason
	n.UpdatedAt = time.Now().UTC()
}
