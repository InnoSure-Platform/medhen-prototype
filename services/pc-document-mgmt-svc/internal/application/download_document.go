package application

import (
	"context"
	"fmt"
	"io"
	"github.com/inno-sphere/medhen-prototype/services/pc-document-mgmt-svc/internal/domain"
)

type DownloadDocumentQuery struct {
	DocumentID string
	RedactPII  bool
}

type DownloadDocumentUseCase struct {
	repo     domain.DocumentRepository
	storage  domain.ObjectStoragePort
	redactor domain.PIIRedactionPort
}

func NewDownloadDocumentUseCase(repo domain.DocumentRepository, storage domain.ObjectStoragePort, redactor domain.PIIRedactionPort) *DownloadDocumentUseCase {
	return &DownloadDocumentUseCase{repo, storage, redactor}
}

func (uc *DownloadDocumentUseCase) Execute(ctx context.Context, query DownloadDocumentQuery) (io.ReadCloser, string, error) {
	doc, err := uc.repo.GetByID(ctx, query.DocumentID)
	if err != nil {
		return nil, "", fmt.Errorf("document not found: %w", err)
	}

	reader, err := uc.storage.DownloadStream(ctx, doc.Storage)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download stream: %w", err)
	}

	if query.RedactPII {
		redactedReader, err := uc.redactor.RedactStream(ctx, reader)
		if err != nil {
			reader.Close()
			return nil, "", fmt.Errorf("failed to apply pii redaction: %w", err)
		}
		return redactedReader, doc.Storage.MimeType, nil
	}

	return reader, doc.Storage.MimeType, nil
}
