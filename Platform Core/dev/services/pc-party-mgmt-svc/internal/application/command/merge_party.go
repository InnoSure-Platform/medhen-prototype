package command

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-party-mgmt-svc/internal/domain"
)

type MergePartyHandler struct {
	uow UnitOfWork
}

func NewMergePartyHandler(uow UnitOfWork) *MergePartyHandler {
	return &MergePartyHandler{
		uow: uow,
	}
}

func (h *MergePartyHandler) Handle(ctx context.Context, cmd MergePartyCommand) error {
	return h.uow.Do(ctx, func(ctx context.Context, repo PartyRepository, outbox OutboxPublisher) error {
		sourceParty, err := repo.FindByID(ctx, cmd.SourcePartyID)
		if err != nil {
			return err
		}

		_, err = repo.FindByID(ctx, cmd.TargetPartyID)
		if err != nil {
			return err
		}

		err = sourceParty.MergeInto(cmd.TargetPartyID)
		if err != nil {
			return err
		}

		if err := repo.Save(ctx, sourceParty); err != nil {
			return err
		}

		event := domain.PartyMergedEvent{
			ID:             uuid.New(),
			TenantID:       cmd.TenantID,
			SourcePartyID:  cmd.SourcePartyID,
			TargetPartyID:  cmd.TargetPartyID,
			MergedBy:       cmd.MergedBy,
			OccurredAtTime: time.Now(),
		}

		return outbox.Publish(ctx, event)
	})
}
