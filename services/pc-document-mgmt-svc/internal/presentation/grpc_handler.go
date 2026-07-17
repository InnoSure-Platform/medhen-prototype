package presentation

import (
	"context"
	"fmt"
	"github.com/inno-sphere/medhen-prototype/services/pc-document-mgmt-svc/internal/application"
)

// DocumentGrpcService simulates the generated gRPC server interface
type DocumentGrpcService struct {
	genUseCase      *application.GenerateDocumentUseCase
	sigUseCase      *application.SignatureWorkflowUseCase
	legalHoldUseCase *application.LegalHoldUseCase
}

func NewDocumentGrpcService(gen *application.GenerateDocumentUseCase, sig *application.SignatureWorkflowUseCase, lh *application.LegalHoldUseCase) *DocumentGrpcService {
	return &DocumentGrpcService{genUseCase: gen, sigUseCase: sig, legalHoldUseCase: lh}
}

// GenerateDocument handler maps proto request to application command
func (s *DocumentGrpcService) GenerateDocument(ctx context.Context, tenantID, templateCode, locale, entityType, entityID string, payload map[string]interface{}) (string, error) {
	cmd := application.GenerateDocumentCommand{
		TenantID:     tenantID,
		TemplateCode: templateCode,
		Locale:       locale,
		EntityType:   entityType,
		EntityID:     entityID,
		Payload:      payload,
	}

	docID, err := s.genUseCase.Execute(ctx, cmd)
	if err != nil {
		return "", fmt.Errorf("failed to handle GenerateDocument command: %w", err)
	}

	return docID, nil
}

func (s *DocumentGrpcService) ApplyLegalHold(ctx context.Context, documentID, reason string) error {
	cmd := application.ApplyLegalHoldCommand{
		DocumentID: documentID,
		Reason:     reason,
	}
	return s.legalHoldUseCase.ApplyHold(ctx, cmd)
}

func (s *DocumentGrpcService) ReleaseLegalHold(ctx context.Context, documentID, reason string) error {
	cmd := application.ReleaseLegalHoldCommand{
		DocumentID: documentID,
		Reason:     reason,
	}
	return s.legalHoldUseCase.ReleaseHold(ctx, cmd)
}
