package commands

import (
	"context"
	"fmt"

	"github.com/medhen/pc-audit-svc/internal/domain/audit"
)

type ExecuteCryptoShreddingCommand struct {
	TenantID       string
	EntityID       string
	ActorID        string
	DEKReferenceID string
}

type ExecuteCryptoShreddingHandler struct {
	ledgerRepo audit.HotLedgerRepository
	kmsService audit.KMSService
}

func NewExecuteCryptoShreddingHandler(ledgerRepo audit.HotLedgerRepository, kmsService audit.KMSService) *ExecuteCryptoShreddingHandler {
	return &ExecuteCryptoShreddingHandler{
		ledgerRepo: ledgerRepo,
		kmsService: kmsService,
	}
}

func (h *ExecuteCryptoShreddingHandler) Handle(ctx context.Context, cmd ExecuteCryptoShreddingCommand) error {
	// 1. Verify no active legal holds exist for the target
	isHeld, err := h.ledgerRepo.CheckLegalHold(ctx, cmd.TenantID, cmd.EntityID, cmd.ActorID)
	if err != nil {
		return fmt.Errorf("failed to check legal hold status: %w", err)
	}
	if isHeld {
		return audit.ErrActiveLegalHold
	}

	// 2. Destroy the Key in KMS
	if err := h.kmsService.DestroyDEK(ctx, cmd.DEKReferenceID); err != nil {
		return fmt.Errorf("failed to destroy DEK in KMS: %w", err)
	}

	// Note: We do NOT delete the record from Postgres or Iceberg.
	// The ciphertext remains, preserving the cryptographic hash chain,
	// but it is permanently undecipherable, satisfying Right-to-be-Forgotten.

	return nil
}
