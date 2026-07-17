package command

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/medhen/pc-billing-svc/internal/application/port"
	"github.com/medhen/pc-billing-svc/internal/domain/aggregate"
	"github.com/shopspring/decimal"
)

type StageAllocationPayload struct {
	AccountID uuid.UUID       `json:"account_id"`
	InvoiceID uuid.UUID       `json:"invoice_id"`
	Amount    decimal.Decimal `json:"amount"`
}

type StageManualAllocationCmd struct {
	TenantID string
	Payload  StageAllocationPayload
	MakerID  string
}

// Staging a manual allocation of suspense funds requires Maker-Checker workflow.
type StageManualAllocationHandler struct {
	uow          port.UnitOfWork
	approvalRepo ApprovalRequestRepository // Requires a port interface addition
}

type ApprovalRequestRepository interface {
	Save(ctx context.Context, request *aggregate.ApprovalRequest) error
}

func NewStageManualAllocationHandler(uow port.UnitOfWork, repo ApprovalRequestRepository) *StageManualAllocationHandler {
	return &StageManualAllocationHandler{
		uow:          uow,
		approvalRepo: repo,
	}
}

func (h *StageManualAllocationHandler) Handle(ctx context.Context, cmd StageManualAllocationCmd) error {
	return h.uow.Execute(ctx, func(ctx context.Context) error {
		if cmd.Payload.Amount.LessThanOrEqual(decimal.Zero) {
			return errors.New("allocation amount must be greater than zero")
		}

		payloadBytes, err := json.Marshal(cmd.Payload)
		if err != nil {
			return err
		}

		approval := aggregate.NewApprovalRequest(
			cmd.TenantID,
			"MANUAL_ALLOCATION",
			cmd.Payload.AccountID,
			payloadBytes,
			cmd.MakerID,
		)

		return h.approvalRepo.Save(ctx, approval)
	})
}
