package application

import (
	"context"
	"fmt"
	"github.com/inno-sphere/medhen-prototype/services/pc-document-mgmt-svc/internal/domain"
)

type ApplyLegalHoldCommand struct {
	DocumentID string
	Reason     string
}

type ReleaseLegalHoldCommand struct {
	DocumentID string
	Reason     string
}

type LegalHoldUseCase struct {
	repo domain.DocumentRepository
}

func NewLegalHoldUseCase(repo domain.DocumentRepository) *LegalHoldUseCase {
	return &LegalHoldUseCase{repo: repo}
}

func (uc *LegalHoldUseCase) ApplyHold(ctx context.Context, cmd ApplyLegalHoldCommand) error {
	doc, err := uc.repo.GetByID(ctx, cmd.DocumentID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	doc.ApplyLegalHold()

	if err := uc.repo.Update(ctx, doc); err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

func (uc *LegalHoldUseCase) ReleaseHold(ctx context.Context, cmd ReleaseLegalHoldCommand) error {
	doc, err := uc.repo.GetByID(ctx, cmd.DocumentID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	doc.ReleaseLegalHold()

	if err := uc.repo.Update(ctx, doc); err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}
