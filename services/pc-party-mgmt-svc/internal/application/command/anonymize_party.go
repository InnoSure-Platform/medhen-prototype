package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-party-mgmt-svc/internal/domain"
)

type AnonymizePartyCommand struct {
	PartyID uuid.UUID
	Reason  string
}

type AnonymizePartyHandler struct {
	uow UnitOfWork
}

func NewAnonymizePartyHandler(uow UnitOfWork) *AnonymizePartyHandler {
	return &AnonymizePartyHandler{uow: uow}
}

func (h *AnonymizePartyHandler) Handle(ctx context.Context, cmd AnonymizePartyCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repo PartyRepository, outbox OutboxPublisher) error {
		party, err := repo.FindByID(ctx, cmd.PartyID)
		if err != nil {
			return err
		}

		if err := party.Anonymize(); err != nil {
			return err
		}

		if err := repo.Save(ctx, party); err != nil {
			return err
		}

		// Prepare domain event
		evt := domain.PartyAnonymizedEvent{
			ID:             uuid.New(),
			TenantID:       party.TenantID,
			PartyID:        party.ID,
			Reason:         cmd.Reason,
			OccurredAtTime: time.Now(),
		}

		if err := outbox.Publish(ctx, evt); err != nil {
			return fmt.Errorf("failed to publish outbox event: %w", err)
		}

		return nil
	})
}
