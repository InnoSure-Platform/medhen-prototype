package grpcclient

import (
	"context"

	"github.com/medhen/pc-underwriting-svc/internal/application/port"
	"github.com/medhen/pc-underwriting-svc/internal/domain/aggregate"
)

type ProductClient struct {
	// grpc connection details would go here
}

func NewProductClient() port.ProductServiceClient {
	return &ProductClient{}
}

func (c *ProductClient) EvaluateDMNRules(ctx context.Context, productID string, riskPayload []byte) (aggregate.AssessmentStatus, int, []string, error) {
	// Mock implementation of DMN execution call to pc-product-defn-svc
	// Industry standard: The DMN engine determines the STP status, the risk score, and lists triggered rules.
	return aggregate.StatusReferred, 45, []string{"RULE_HIGH_TSI", "RULE_PRIOR_CLAIMS"}, nil
}
