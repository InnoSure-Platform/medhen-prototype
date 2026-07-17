package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/inno-sphere/medhen-prototype/services/pc-document-mgmt-svc/internal/domain"
	"time"
)

type GenerateDocumentCommand struct {
	TenantID     string
	TemplateCode string
	Locale       string
	EntityType   string
	EntityID     string
	Payload      map[string]interface{}
}

type GenerateDocumentUseCase struct {
	repo       domain.DocumentRepository
	tmplRepo   domain.TemplateRepository
	storage    domain.ObjectStoragePort
	renderer   domain.DocumentRendererPort
	publisher  domain.EventPublisherPort
}

func NewGenerateDocumentUseCase(
	repo domain.DocumentRepository,
	tmplRepo domain.TemplateRepository,
	storage domain.ObjectStoragePort,
	renderer domain.DocumentRendererPort,
	publisher domain.EventPublisherPort,
) *GenerateDocumentUseCase {
	return &GenerateDocumentUseCase{repo, tmplRepo, storage, renderer, publisher}
}

func (uc *GenerateDocumentUseCase) Execute(ctx context.Context, cmd GenerateDocumentCommand) (string, error) {
	// 1. Fetch Template
	tmpl, err := uc.tmplRepo.GetActiveByCodeAndLocale(ctx, cmd.TemplateCode, cmd.Locale)
	if err != nil {
		return "", fmt.Errorf("template resolution failed: %w", err)
	}

	// 2. Validate Schema (Simplified for stub)
	// In production, use gojsonschema against tmpl.MergeSchema

	// 3. Render PDF and HTML
	pdfBytes, htmlContent, err := uc.renderer.Render(ctx, tmpl, cmd.Payload)
	if err != nil {
		return "", fmt.Errorf("pdf rendering failed: %w", err)
	}

	// 4. Calculate SHA256 (PDF is the canonical artifact)
	hash := sha256.Sum256(pdfBytes)
	hashString := hex.EncodeToString(hash[:])

	// 5. Upload to Ozone
	docID := uuid.New().String()
	bucket := "active-documents"
	pdfPath := fmt.Sprintf("/%s/%d/%02d/%s/%s.pdf", cmd.EntityType, time.Now().Year(), time.Now().Month(), cmd.EntityID, docID)
	htmlPath := fmt.Sprintf("/%s/%d/%02d/%s/%s.html", cmd.EntityType, time.Now().Year(), time.Now().Month(), cmd.EntityID, docID)
	
	ref, err := uc.storage.UploadStream(ctx, bucket, pdfPath, bytes.NewReader(pdfBytes), int64(len(pdfBytes)), "application/pdf")
	if err != nil {
		return "", fmt.Errorf("pdf upload failed: %w", err)
	}

	htmlBytes := []byte(htmlContent)
	_, err = uc.storage.UploadStream(ctx, bucket, htmlPath, bytes.NewReader(htmlBytes), int64(len(htmlBytes)), "text/html")
	if err != nil {
		// Log error, but don't fail the entire generation if HTML fails, or decide to fail:
		return "", fmt.Errorf("html upload failed: %w", err)
	}

	// 6. Create Domain Record
	record := domain.NewDocumentRecord(
		docID, cmd.TenantID, cmd.TemplateCode, cmd.EntityType, cmd.EntityID, cmd.Locale,
		domain.StatusActive, int64(len(pdfBytes)), hashString, ref, &htmlPath,
	)

	// 7. Save and Publish Event (Should be transactional outbox in infra)
	if err := uc.repo.Save(ctx, record); err != nil {
		return "", err
	}

	event := domain.DocumentGeneratedEvent{
		BaseEvent: domain.BaseEvent{ID: uuid.New().String(), Timestamp: time.Now().UTC()},
		TenantID:     record.TenantID,
		DocumentID:   record.ID,
		DocumentType: record.DocumentType,
		EntityType:   record.EntityType,
		EntityID:     record.EntityID,
		Locale:       record.Locale,
		StorageURI:   ref.Path, // Simplified URI
		SHA256Hash:   record.SHA256Hash,
	}

	if err := uc.publisher.Publish(ctx, event); err != nil {
		// Log error, rely on outbox relay for actual durability
	}

	return docID, nil
}
