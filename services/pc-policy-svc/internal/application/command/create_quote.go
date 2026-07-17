package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/medhen/pc-policy-svc/internal/domain/quote"
	"github.com/medhen/pc-policy-svc/internal/infrastructure/grpc_client"
)

type CreateQuoteCommand struct {
	TenantID    string
	ProductID   uuid.UUID
	PartyID     uuid.UUID
	RiskPayload []byte
}

type CreateQuoteHandler struct {
	quoteRepo quote.Repository
	uwClient  *grpc_client.UWClient
	// In a real app we'd inject rating client here too
}

func NewCreateQuoteHandler(quoteRepo quote.Repository, uwClient *grpc_client.UWClient) *CreateQuoteHandler {
	return &CreateQuoteHandler{
		quoteRepo: quoteRepo,
		uwClient:  uwClient,
	}
}

func (h *CreateQuoteHandler) Handle(ctx context.Context, cmd CreateQuoteCommand) (*quote.Quote, error) {
	q := quote.NewQuote(cmd.TenantID, cmd.ProductID, cmd.PartyID, cmd.RiskPayload)

	// Check Underwriting Rules
	if h.uwClient != nil {
		uwStatus, err := h.uwClient.AssessQuote(ctx, cmd.TenantID, q.ID, cmd.ProductID.String(), cmd.RiskPayload)
		if err != nil {
			return nil, err
		}
		if uwStatus == "REFERRED" || uwStatus == "DECLINED" {
			q.Status = quote.Status(uwStatus)
		} else {
			q.Calculate(500.00, false, 30) // Mocked Rating Call
		}
	} else {
		q.Calculate(500.00, false, 30)
	}

	err := h.quoteRepo.Save(ctx, q)
	if err != nil {
		return nil, err
	}

	return q, nil
}
