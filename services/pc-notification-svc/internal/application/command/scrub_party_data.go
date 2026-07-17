package command

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"pc-notification-svc/internal/domain/notification"
)

type ScrubPartyDataCommand struct {
	PartyID uuid.UUID
}

type ScrubPartyDataHandler struct {
	repo notification.NotificationRepository
}

func NewScrubPartyDataHandler(repo notification.NotificationRepository) *ScrubPartyDataHandler {
	return &ScrubPartyDataHandler{repo: repo}
}

func (h *ScrubPartyDataHandler) Handle(ctx context.Context, cmd ScrubPartyDataCommand) error {
	// Execute the Right to be Forgotten
	// This will anonymize PII in the notifications table while retaining the metadata row
	err := h.repo.ScrubPartyData(ctx, cmd.PartyID)
	if err != nil {
		return fmt.Errorf("failed to scrub party data: %w", err)
	}
	return nil
}
