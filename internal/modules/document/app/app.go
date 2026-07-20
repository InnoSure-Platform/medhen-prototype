// Package app holds the document use cases.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/document/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/outbox"
)

// ErrNotFound is returned when a document does not exist.
var ErrNotFound = errors.New("document: not found")

// Repository persists documents.
type Repository interface {
	Save(ctx context.Context, d *domain.Document) error
	Get(ctx context.Context, tenantID, id string) (*domain.Document, error)
	GetByPolicy(ctx context.Context, tenantID, policyID string) (*domain.Document, error)
}

// Service implements document use cases.
type Service struct {
	db   *database.DB
	repo Repository
}

// NewService builds the service.
func NewService(db *database.DB, repo Repository) *Service { return &Service{db: db, repo: repo} }

// GenerateCertificate creates and stores a COI for a policy (idempotent per
// policy), emitting document.generated in the same transaction.
func (s *Service) GenerateCertificate(ctx context.Context, tenantID, policyID, policyNumber, partyName string) (*domain.Document, error) {
	if existing, err := s.repo.GetByPolicy(ctx, tenantID, policyID); err == nil {
		return existing, nil
	}

	doc := domain.NewCertificate(tenantID, policyID, policyNumber, partyName)
	err := s.db.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.repo.Save(ctx, doc); err != nil {
			return err
		}
		evt := domain.DocumentGenerated{
			DocumentID: doc.ID, TenantID: tenantID, PolicyID: policyID,
			Type: doc.Type, OccurredAt: time.Now().UTC(),
		}
		payload, err := json.Marshal(evt)
		if err != nil {
			return fmt.Errorf("document: marshal event: %w", err)
		}
		return outbox.Write(ctx, s.db.Conn(ctx), outbox.Message{
			ID: ids.New(), Topic: domain.TopicDocumentGenerated,
			AggregateType: "document", AggregateID: doc.ID, Payload: payload,
		})
	})
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// Get loads a document.
func (s *Service) Get(ctx context.Context, tenantID, id string) (*domain.Document, error) {
	return s.repo.Get(ctx, tenantID, id)
}
