package command

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"github.com/medhen/pc-underwriting-svc/internal/application/port"
	"github.com/medhen/pc-underwriting-svc/internal/domain/valueobject"
)

type SubmitDecisionCommand struct {
	TenantID           string
	ReferralID         uuid.UUID
	UnderwriterID      string
	Decision           valueobject.DecisionType
	Conditions         []valueobject.Condition
	Disclosures        []string
	ActorLevelCode     string
	ProductLOB         string
	Premium            float64
	TSI                float64
	FacultativeCleared bool
}

type SubmitDecisionHandler struct {
	uow           port.UnitOfWork
	referralRepo  port.ReferralRepository
	authorityRepo port.AuthorityRepository
	outboxRepo    port.OutboxRepository
}

func NewSubmitDecisionHandler(uow port.UnitOfWork, rr port.ReferralRepository, ar port.AuthorityRepository, or port.OutboxRepository) *SubmitDecisionHandler {
	return &SubmitDecisionHandler{
		uow:           uow,
		referralRepo:  rr,
		authorityRepo: ar,
		outboxRepo:    or,
	}
}

func (h *SubmitDecisionHandler) Handle(ctx context.Context, cmd SubmitDecisionCommand) error {
	ctx, span := otel.Tracer("pc-underwriting-svc").Start(ctx, "SubmitDecisionHandler.Handle")
	defer span.End()

	return h.uow.Do(ctx, func(txCtx context.Context) error {
		referral, err := h.referralRepo.FindByID(txCtx, cmd.ReferralID)
		if err != nil {
			return err
		}

		authority, err := h.authorityRepo.FindByLevelAndProduct(txCtx, cmd.TenantID, cmd.ActorLevelCode, cmd.ProductLOB)
		if err != nil {
			return err
		}

		err = referral.Decide(cmd.Decision, cmd.Conditions, cmd.Disclosures, authority, cmd.Premium, cmd.TSI, cmd.FacultativeCleared)
		if err != nil {
			return err
		}

		if err := h.referralRepo.Update(txCtx, referral); err != nil {
			return err
		}

		// Emit ReferralDecided event
		payload, _ := json.Marshal(map[string]interface{}{
			"referral_id":     referral.ID,
			"decision":        cmd.Decision,
			"underwriter_id":  cmd.UnderwriterID,
			"authority_level": cmd.ActorLevelCode,
		})
		if err := h.outboxRepo.PublishEvent(txCtx, "pc.underwriting.referral.decided.v1", referral.AssessmentID.String(), payload); err != nil {
			return err
		}

		return nil
	})
}
