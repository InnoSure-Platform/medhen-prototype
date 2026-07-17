package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/medhen/pc-underwriting-svc/internal/application/port"
)

type AssignReferralCommand struct {
	TenantID      string
	ReferralID    uuid.UUID
	UnderwriterID string
}

type AssignReferralHandler struct {
	uow          port.UnitOfWork
	referralRepo port.ReferralRepository
}

func NewAssignReferralHandler(uow port.UnitOfWork, rr port.ReferralRepository) *AssignReferralHandler {
	return &AssignReferralHandler{
		uow:          uow,
		referralRepo: rr,
	}
}

func (h *AssignReferralHandler) Handle(ctx context.Context, cmd AssignReferralCommand) error {
	return h.uow.Do(ctx, func(txCtx context.Context) error {
		referral, err := h.referralRepo.FindByID(txCtx, cmd.ReferralID)
		if err != nil {
			return err
		}

		if err := referral.Assign(cmd.UnderwriterID); err != nil {
			return err
		}

		return h.referralRepo.Update(txCtx, referral)
	})
}
