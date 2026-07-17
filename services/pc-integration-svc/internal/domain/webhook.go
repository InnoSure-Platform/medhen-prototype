package domain

import (
	"time"

	"github.com/google/uuid"
)

type WebhookStatus string

const (
	WebhookStatusProcessed        WebhookStatus = "PROCESSED"
	WebhookStatusIgnoredDuplicate WebhookStatus = "IGNORED_DUPLICATE"
	WebhookStatusFailedSignature  WebhookStatus = "FAILED_SIGNATURE"
)

// WebhookReceipt tracks received callbacks for idempotency and auditing.
type WebhookReceipt struct {
	ID                    uuid.UUID
	Provider              string
	ProviderTransactionID string
	Status                WebhookStatus
	RawPayloadEncrypted   []byte
	ReceivedAt            time.Time
}

// NewWebhookReceipt initializes a tracking record for a webhook.
func NewWebhookReceipt(provider, providerTxnID string, rawPayload []byte, status WebhookStatus) *WebhookReceipt {
	return &WebhookReceipt{
		ID:                    uuid.New(),
		Provider:              provider,
		ProviderTransactionID: providerTxnID,
		Status:                status,
		RawPayloadEncrypted:   rawPayload, // In a real system, encrypt this
		ReceivedAt:            time.Now(),
	}
}
