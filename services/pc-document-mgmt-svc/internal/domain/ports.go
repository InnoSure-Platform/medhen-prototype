package domain

import (
	"context"
	"io"
)

type DocumentRepository interface {
	Save(ctx context.Context, doc *DocumentRecord) error
	GetByID(ctx context.Context, id string) (*DocumentRecord, error)
	Update(ctx context.Context, doc *DocumentRecord) error
}

type TemplateRepository interface {
	Save(ctx context.Context, template *DocumentTemplate) error
	GetByCodeAndVersion(ctx context.Context, code string, version int) (*DocumentTemplate, error)
	GetActiveByCodeAndLocale(ctx context.Context, code, locale string) (*DocumentTemplate, error)
}

type SignatureRepository interface {
	Save(ctx context.Context, req *SignatureRequest) error
	GetByID(ctx context.Context, id string) (*SignatureRequest, error)
	Update(ctx context.Context, req *SignatureRequest) error
}

type EventPublisherPort interface {
	Publish(ctx context.Context, event Event) error
}

type ObjectStoragePort interface {
	UploadStream(ctx context.Context, bucket, path string, reader io.Reader, size int64, mimeType string) (StorageRef, error)
	DownloadStream(ctx context.Context, ref StorageRef) (io.ReadCloser, error)
}

type MalwareScannerPort interface {
	ScanStream(ctx context.Context, reader io.Reader) (bool, error) // Returns true if clean
}

type PIIRedactionPort interface {
	RedactStream(ctx context.Context, reader io.Reader) (io.ReadCloser, error)
}

type DocumentRendererPort interface {
	Render(ctx context.Context, template *DocumentTemplate, payload map[string]interface{}) (pdfBytes []byte, htmlContent string, err error)
}
