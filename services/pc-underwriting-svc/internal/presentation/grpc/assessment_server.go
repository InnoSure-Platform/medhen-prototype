package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/medhen/pc-underwriting-svc/internal/application/command"
)

// AssessmentServer implements the gRPC interface (mocked struct)
type AssessmentServer struct {
	assessRiskHandler *command.AssessRiskHandler
}

func NewAssessmentServer(arh *command.AssessRiskHandler) *AssessmentServer {
	return &AssessmentServer{
		assessRiskHandler: arh,
	}
}

// AssessQuote simulates the gRPC AssessQuote RPC
func (s *AssessmentServer) AssessQuote(ctx context.Context, tenantID string, quoteID uuid.UUID, productID string, riskPayload []byte) (string, error) {
	cmd := command.AssessRiskCommand{
		TenantID:    tenantID,
		QuoteID:     quoteID,
		ProductID:   productID,
		ProductLOB:  "motor", // Simplified
		RiskPayload: riskPayload,
		Premium:     1000.0,
		TSI:         100000.0,
	}

	assessment, err := s.assessRiskHandler.Handle(ctx, cmd)
	if err != nil {
		return "", err
	}

	return string(assessment.Status), nil
}
