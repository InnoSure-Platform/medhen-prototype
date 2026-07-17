package application

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/inno-sphere/medhen-prototype/services/pc-document-mgmt-svc/internal/domain"
	"time"
)

type InitiateSignatureCommand struct {
	DocumentID string
	TenantID   string
	Signatory  string
}

type SignDocumentCommand struct {
	SignatureID string
	IPAddress   string
	UserAgent   string
}

type SignatureWorkflowUseCase struct {
	docRepo domain.DocumentRepository
	sigRepo domain.SignatureRepository
	pub     domain.EventPublisherPort
}

func NewSignatureWorkflowUseCase(docRepo domain.DocumentRepository, sigRepo domain.SignatureRepository, pub domain.EventPublisherPort) *SignatureWorkflowUseCase {
	return &SignatureWorkflowUseCase{docRepo, sigRepo, pub}
}

func (uc *SignatureWorkflowUseCase) InitiateSignature(ctx context.Context, cmd InitiateSignatureCommand) (string, error) {
	doc, err := uc.docRepo.GetByID(ctx, cmd.DocumentID)
	if err != nil {
		return "", err
	}
	if doc.Status != domain.StatusActive {
		return "", domain.ErrInvalidState
	}

	sigID := uuid.New().String()
	req := domain.NewSignatureRequest(sigID, cmd.DocumentID, cmd.TenantID, cmd.Signatory)
	if err := uc.sigRepo.Save(ctx, req); err != nil {
		return "", err
	}

	return sigID, nil
}

func (uc *SignatureWorkflowUseCase) SignDocument(ctx context.Context, cmd SignDocumentCommand) error {
	req, err := uc.sigRepo.GetByID(ctx, cmd.SignatureID)
	if err != nil {
		return err
	}

	if err := req.Sign(cmd.IPAddress, cmd.UserAgent); err != nil {
		return err
	}

	if err := uc.sigRepo.Update(ctx, req); err != nil {
		return err
	}

	// Update document signature status
	doc, err := uc.docRepo.GetByID(ctx, req.DocumentID)
	if err == nil {
		status := string(domain.SigStateSigned)
		doc.SignatureStatus = &status
		_ = uc.docRepo.Update(ctx, doc)
	}

	event := domain.DocumentSignedEvent{
		BaseEvent:  domain.BaseEvent{ID: uuid.New().String(), Timestamp: time.Now().UTC()},
		TenantID:   req.TenantID,
		DocumentID: req.DocumentID,
	}
	return uc.pub.Publish(ctx, event)
}
