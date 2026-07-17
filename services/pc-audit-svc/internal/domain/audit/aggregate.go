package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrHashChainInvalid = errors.New("cryptographic validation failed")
	ErrActiveLegalHold  = errors.New("operation violates an active legal hold")
	ErrInvalidSignature = errors.New("digital signature verification failed")
)

// MerkleTreeRoot tracks the immutable top hash of an epoch.
type MerkleTreeRoot struct {
	EpochID           int64
	RootHash          string
	PublishedToNotary bool
	CreatedAt         time.Time
}

// AuditLedgerEntry is the core aggregate root for an immutable audit record (a leaf in the Merkle Tree).
type AuditLedgerEntry struct {
	SeqID      int64
	EventID    uuid.UUID
	TenantID   string
	Timestamp  time.Time
	Actor      ActorContext
	ActionType string
	EntityType string
	EntityID   string

	// Provenance & Integrity (Phase 1 & 2)
	TraceID          string
	ProducerKeyID    string
	DigitalSignature []byte

	IsPIIEncrypted bool
	CryptoEnvelope *CryptoEnvelope
	DeltaPlaintext []byte

	MerkleLeafHash string
}

// NewAuditLedgerEntry creates a new entry and computes its leaf hash for the Merkle tree.
func NewAuditLedgerEntry(
	seqID int64,
	tenantID string,
	actor ActorContext,
	actionType, entityType, entityID string,
	traceID, producerKeyID string,
	digitalSignature []byte,
	isPIIEncrypted bool,
	envelope *CryptoEnvelope,
	deltaPlaintext []byte,
) (*AuditLedgerEntry, error) {
	entry := &AuditLedgerEntry{
		SeqID:            seqID,
		EventID:          uuid.New(),
		TenantID:         tenantID,
		Timestamp:        time.Now().UTC(),
		Actor:            actor,
		ActionType:       actionType,
		EntityType:       entityType,
		EntityID:         entityID,
		TraceID:          traceID,
		ProducerKeyID:    producerKeyID,
		DigitalSignature: digitalSignature,
		IsPIIEncrypted:   isPIIEncrypted,
		CryptoEnvelope:   envelope,
		DeltaPlaintext:   deltaPlaintext,
	}

	hash, err := entry.computeLeafHash()
	if err != nil {
		return nil, err
	}
	entry.MerkleLeafHash = hash

	return entry, nil
}

// computeLeafHash computes the SHA-256 hash forming the leaf node of the Merkle Tree.
func (e *AuditLedgerEntry) computeLeafHash() (string, error) {
	h := sha256.New()

	// Add critical immutable fields
	h.Write([]byte(e.EventID.String()))
	h.Write([]byte(e.TenantID))
	h.Write([]byte(e.Actor.UserID))
	h.Write([]byte(e.ActionType))
	h.Write([]byte(e.EntityID))
	h.Write([]byte(e.TraceID))
	h.Write(e.DigitalSignature)

	// Add payload (either ciphertext or plaintext)
	if e.IsPIIEncrypted && e.CryptoEnvelope != nil {
		h.Write(e.CryptoEnvelope.Ciphertext)
	} else if e.DeltaPlaintext != nil {
		h.Write(e.DeltaPlaintext)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ShredPII implements the Right to be Forgotten by nullifying the DEK reference.
// Note: This does NOT alter the MerkleLeafHash, as the cipher remains identical.
func (e *AuditLedgerEntry) ShredPII() {
	if e.IsPIIEncrypted && e.CryptoEnvelope != nil {
		e.CryptoEnvelope.DEKReferenceID = "" // Shredded
	}
}

// ExportJob tracks the lifecycle of massive data exports.
type ExportJob struct {
	ID        uuid.UUID
	TenantID  string
	Status    string // QUEUED, PROCESSING, PACKAGING, COMPLETED, FAILED
	Query     string
	S3URL     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// LegalHold represents a compliance freeze on data retention rules.
type LegalHold struct {
	ID             uuid.UUID
	TenantID       string
	TargetActorID  string
	TargetEntityID string
	Justification  string
	IsActive       bool
	CreatedAt      time.Time
}
