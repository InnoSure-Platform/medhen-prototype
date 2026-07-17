package domain

import (
	"time"
)

type DocumentStatus string

const (
	StatusGenerating   DocumentStatus = "GENERATING"
	StatusPendingScan  DocumentStatus = "PENDING_SCAN"
	StatusVerified     DocumentStatus = "VERIFIED"
	StatusQuarantined  DocumentStatus = "QUARANTINED"
	StatusProcessingIDP DocumentStatus = "PROCESSING_IDP"
	StatusActive       DocumentStatus = "ACTIVE"
	StatusArchived     DocumentStatus = "ARCHIVED"
)

type StorageRef struct {
	Volume   string
	Bucket   string
	Path     string
	MimeType string
}

type ExtractionResult struct {
	ExtractedData   map[string]interface{}
	ConfidenceScore float64
	RequiresReview  bool
}

type DocumentRecord struct {
	ID               string
	TenantID         string
	DocumentType     string
	EntityType       string
	EntityID         string
	Locale           string
	Status           DocumentStatus
	FileSize         int64
	SHA256Hash       string
	Storage          StorageRef
	HtmlStoragePath  *string
	IsLegalHold      bool
	IDPExtractedData *ExtractionResult
	SignatureStatus  *string
	CreatedAt        time.Time
}

func NewDocumentRecord(id, tenantID, docType, entityType, entityID, locale string, status DocumentStatus, size int64, hash string, storage StorageRef, htmlPath *string) *DocumentRecord {
	return &DocumentRecord{
		ID:              id,
		TenantID:        tenantID,
		DocumentType:    docType,
		EntityType:      entityType,
		EntityID:        entityID,
		Locale:          locale,
		Status:          status,
		FileSize:        size,
		SHA256Hash:      hash,
		Storage:         storage,
		HtmlStoragePath: htmlPath,
		IsLegalHold:     false,
		CreatedAt:       time.Now().UTC(),
	}
}

func (d *DocumentRecord) MarkVerified() error {
	if d.Status != StatusPendingScan {
		return ErrInvalidState
	}
	d.Status = StatusVerified
	return nil
}

func (d *DocumentRecord) MarkQuarantined() error {
	if d.Status != StatusPendingScan {
		return ErrInvalidState
	}
	d.Status = StatusQuarantined
	return nil
}

func (d *DocumentRecord) MarkActive() error {
	if d.Status != StatusGenerating && d.Status != StatusVerified && d.Status != StatusProcessingIDP {
		return ErrInvalidState
	}
	d.Status = StatusActive
	return nil
}

func (d *DocumentRecord) MarkArchived() error {
	if d.IsLegalHold {
		return ErrLegalHoldActive
	}
	d.Status = StatusArchived
	return nil
}

func (d *DocumentRecord) ApplyLegalHold() {
	d.IsLegalHold = true
}

func (d *DocumentRecord) ReleaseLegalHold() {
	d.IsLegalHold = false
}

func (d *DocumentRecord) AttachIDPResult(res ExtractionResult) {
	d.IDPExtractedData = &res
	d.Status = StatusActive
}
