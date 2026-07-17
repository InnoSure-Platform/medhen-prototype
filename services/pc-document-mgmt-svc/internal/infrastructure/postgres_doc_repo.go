package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/inno-sphere/medhen-prototype/services/pc-document-mgmt-svc/internal/domain"
)

type PostgresDocumentRepository struct {
	db *sql.DB
}

func NewPostgresDocumentRepository(db *sql.DB) *PostgresDocumentRepository {
	return &PostgresDocumentRepository{db: db}
}

func (r *PostgresDocumentRepository) Save(ctx context.Context, doc *domain.DocumentRecord) error {
	query := `
		INSERT INTO document_records (
			id, tenant_id, document_type, entity_type, entity_id, locale, status, mime_type, file_size_bytes, sha256_hash, storage_volume, storage_bucket, storage_path, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	
	_, err := r.db.ExecContext(ctx, query,
		doc.ID, doc.TenantID, doc.DocumentType, doc.EntityType, doc.EntityID, doc.Locale, doc.Status,
		doc.Storage.MimeType, doc.FileSize, doc.SHA256Hash, doc.Storage.Volume, doc.Storage.Bucket, doc.Storage.Path, doc.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save document record: %w", err)
	}
	return nil
}

func (r *PostgresDocumentRepository) GetByID(ctx context.Context, id string) (*domain.DocumentRecord, error) {
	// Stub implementation for compilation
	return nil, domain.ErrDocumentNotFound
}

func (r *PostgresDocumentRepository) Update(ctx context.Context, doc *domain.DocumentRecord) error {
	query := `UPDATE document_records SET status = $1, signature_status = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, doc.Status, doc.SignatureStatus, doc.ID)
	return err
}
