package aggregate

import (
	"time"

	"github.com/google/uuid"
)

type MandateStatus string

const (
	MandateStatusActive    MandateStatus = "ACTIVE"
	MandateStatusRevoked   MandateStatus = "REVOKED"
	MandateStatusSuspended MandateStatus = "SUSPENDED"
)

// PaymentMandate securely stores tokenized references for auto-debit collections.
// The raw PAN/Wallet info is stored upstream in the Integration ACL / Gateway.
type PaymentMandate struct {
	ID               uuid.UUID
	TenantID         string
	BillingAccountID uuid.UUID
	Provider         string // e.g., "TELEBIRR", "CBE", "CYBERSOURCE"
	ProviderToken    string // The secure vault token
	Status           MandateStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewPaymentMandate(tenantID string, accountID uuid.UUID, provider, token string) *PaymentMandate {
	return &PaymentMandate{
		ID:               uuid.New(),
		TenantID:         tenantID,
		BillingAccountID: accountID,
		Provider:         provider,
		ProviderToken:    token,
		Status:           MandateStatusActive,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
}

func (m *PaymentMandate) Revoke() {
	m.Status = MandateStatusRevoked
	m.UpdatedAt = time.Now().UTC()
}
