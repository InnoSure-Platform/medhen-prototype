package commands

import (
	"context"

	"github.com/shopspring/decimal"
	"medhen/pc-claims-svc/internal/domain"
)

type AdjustReserveCommand struct {
	ClaimID         string
	ReserveType     domain.ReserveType
	TransactionType string
	Amount          domain.MultiCurrencyAmount
	ReasonCode      string
	AuthorID        string
	AuthorityLimit  decimal.Decimal
}

type ReserveRepository interface {
	GetLedger(ctx context.Context, claimID string) (*domain.ReserveLedger, error)
	SaveLedger(ctx context.Context, ledger *domain.ReserveLedger, eventPayload []byte) error
}

type AdjustReserveHandler struct {
	repo ReserveRepository
}

func NewAdjustReserveHandler(repo ReserveRepository) *AdjustReserveHandler {
	return &AdjustReserveHandler{repo: repo}
}

func (h *AdjustReserveHandler) Handle(ctx context.Context, cmd AdjustReserveCommand) error {
	ledger, err := h.repo.GetLedger(ctx, cmd.ClaimID)
	if err != nil {
		return err
	}

	entry := domain.ReserveEntry{
		ClaimID:         cmd.ClaimID,
		ReserveType:     cmd.ReserveType,
		TransactionType: cmd.TransactionType,
		Amount:          cmd.Amount,
		ReasonCode:      cmd.ReasonCode,
		AuthorID:        cmd.AuthorID,
	}

	err = ledger.AdjustReserve(entry, cmd.AuthorityLimit)
	if err != nil {
		return err
	}

	// Avro Mock Payload
	eventPayload := []byte(`{"event_id": "mock", "reserve_type": "` + string(cmd.ReserveType) + `"}`)

	return h.repo.SaveLedger(ctx, ledger, eventPayload)
}
