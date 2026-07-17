package command

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-policy-svc/internal/domain/policy"
	"github.com/medhen/pc-policy-svc/internal/infrastructure/grpc_client"
)

type EndorsePolicyCommand struct {
	PolicyID      uuid.UUID
	EffectiveDate time.Time
	RiskPayload   []byte
	NewPremium    float64
}

type EndorsePolicyHandler struct {
	policyRepo policy.Repository
	uwClient   *grpc_client.UWClient
}

func NewEndorsePolicyHandler(policyRepo policy.Repository, uwClient *grpc_client.UWClient) *EndorsePolicyHandler {
	return &EndorsePolicyHandler{
		policyRepo: policyRepo,
		uwClient:   uwClient,
	}
}

func (h *EndorsePolicyHandler) Handle(ctx context.Context, cmd EndorsePolicyCommand) (*policy.PolicyVersion, error) {
	p, err := h.policyRepo.GetByID(ctx, cmd.PolicyID)
	if err != nil {
		return nil, err
	}

	// UW Check for mid-term adjustments
	if h.uwClient != nil {
		uwStatus, err := h.uwClient.AssessQuote(ctx, p.TenantID, uuid.Nil, p.ProductID.String(), cmd.RiskPayload)
		if err != nil {
			return nil, err
		}
		if uwStatus == "REFERRED" || uwStatus == "DECLINED" {
			// In a real flow, this would go into a pending endorsement queue
			return nil, errors.New("endorsement " + uwStatus + " by underwriting rules")
		}
	}

	newVer, err := p.Endorse(cmd.EffectiveDate, cmd.RiskPayload, cmd.NewPremium)
	if err != nil {
		return nil, err
	}

	err = h.policyRepo.Save(ctx, p)
	if err != nil {
		return nil, err
	}

	return newVer, nil
}
