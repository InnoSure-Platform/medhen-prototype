package domain

import (
	"time"

	"github.com/google/uuid"
)

type KYCStatus string

const (
	KYCStatusPending  KYCStatus = "PENDING"
	KYCStatusVerified KYCStatus = "VERIFIED"
	KYCStatusRejected KYCStatus = "REJECTED"
	KYCStatusExpired  KYCStatus = "EXPIRED"
)

type KYCDocument struct {
	ID         uuid.UUID
	PartyID    uuid.UUID
	Type       string // PASSPORT, NATIONAL_ID, BUSINESS_LICENSE
	ObjectKey  string // MinIO object key
	Status     KYCStatus
	IssueDate  *time.Time
	ExpiryDate *time.Time
	VerifiedBy string
	VerifiedAt *time.Time
	CreatedAt  time.Time
}

func NewKYCDocument(partyID uuid.UUID, docType, objectKey string) *KYCDocument {
	return &KYCDocument{
		ID:        uuid.New(),
		PartyID:   partyID,
		Type:      docType,
		ObjectKey: objectKey,
		Status:    KYCStatusPending,
		CreatedAt: time.Now(),
	}
}

func (d *KYCDocument) Verify(verifierID string) {
	d.Status = KYCStatusVerified
	d.VerifiedBy = verifierID
	now := time.Now()
	d.VerifiedAt = &now
}

func (d *KYCDocument) Reject(verifierID string) {
	d.Status = KYCStatusRejected
	d.VerifiedBy = verifierID
	now := time.Now()
	d.VerifiedAt = &now
}
