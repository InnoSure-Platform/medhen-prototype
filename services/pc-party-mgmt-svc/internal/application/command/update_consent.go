package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-party-mgmt-svc/internal/domain"
)

type UpdateConsentCommand struct {
	PartyID     uuid.UUID
	ConsentType string
	Status      domain.ConsentStatus
}

type UpdateConsentHandler struct {
	uow UnitOfWork
}

func NewUpdateConsentHandler(uow UnitOfWork) *UpdateConsentHandler {
	return &UpdateConsentHandler{uow: uow}
}

func (h *UpdateConsentHandler) Handle(ctx context.Context, cmd UpdateConsentCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repo PartyRepository, outbox OutboxPublisher) error {
		party, err := repo.FindByID(ctx, cmd.PartyID)
		if err != nil {
			return err
		}

		party.UpdateConsent(cmd.ConsentType, cmd.Status)

		if err := repo.Save(ctx, party); err != nil {
			return err
		}

		// Prepare domain event
		evt := domain.PartyConsentUpdatedEvent{
			ID:             uuid.New(),
			TenantID:       party.TenantID,
			PartyID:        party.ID,
			ConsentType:    cmd.ConsentType,
			Status:         string(cmd.Status),
			Version:        party.Version,
			OccurredAtTime: time.Now(),
		}

		if err := outbox.Publish(ctx, evt); err != nil {
			return fmt.Errorf("failed to publish outbox event: %w", err)
		}

		return nil
	})
}
