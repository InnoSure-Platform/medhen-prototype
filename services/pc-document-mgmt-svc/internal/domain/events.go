package domain

import "time"

// Event is the base interface for domain events
type Event interface {
	EventID() string
	OccurredAt() time.Time
	Topic() string
	PartitionKey() string
}

type BaseEvent struct {
	ID        string    `json:"event_id"`
	Timestamp time.Time `json:"occurred_at"`
}

func (b BaseEvent) EventID() string       { return b.ID }
func (b BaseEvent) OccurredAt() time.Time { return b.Timestamp }

// DocumentGeneratedEvent
type DocumentGeneratedEvent struct {
	BaseEvent
	TenantID     string `json:"tenant_id"`
	DocumentID   string `json:"document_id"`
	DocumentType string `json:"document_type"`
	EntityType   string `json:"entity_type"`
	EntityID     string `json:"entity_id"`
	Locale       string `json:"locale"`
	StorageURI   string `json:"storage_uri"`
	SHA256Hash   string `json:"sha256_hash"`
}

func (e DocumentGeneratedEvent) Topic() string        { return "platform.document.generated.v1" }
func (e DocumentGeneratedEvent) PartitionKey() string { return e.TenantID + ":" + e.EntityID }

// DocumentUploadedEvent
type DocumentUploadedEvent struct {
	BaseEvent
	TenantID   string `json:"tenant_id"`
	DocumentID string `json:"document_id"`
	EntityID   string `json:"entity_id"`
}

func (e DocumentUploadedEvent) Topic() string        { return "platform.document.uploaded.v1" }
func (e DocumentUploadedEvent) PartitionKey() string { return e.TenantID + ":" + e.EntityID }

// DocumentSignedEvent
type DocumentSignedEvent struct {
	BaseEvent
	TenantID   string `json:"tenant_id"`
	DocumentID string `json:"document_id"`
}

func (e DocumentSignedEvent) Topic() string        { return "platform.document.signature.signed.v1" }
func (e DocumentSignedEvent) PartitionKey() string { return e.TenantID + ":" + e.DocumentID }

// IDPExtractionCompleteEvent
type IDPExtractionCompleteEvent struct {
	BaseEvent
	TenantID   string `json:"tenant_id"`
	DocumentID string `json:"document_id"`
}

func (e IDPExtractionCompleteEvent) Topic() string        { return "platform.document.idp.completed.v1" }
func (e IDPExtractionCompleteEvent) PartitionKey() string { return e.TenantID + ":" + e.DocumentID }
