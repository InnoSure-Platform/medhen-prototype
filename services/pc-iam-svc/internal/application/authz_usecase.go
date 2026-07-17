package application

import (
	"context"
	"fmt"

	iamv1 "github.com/medhen/pc-contracts/gen/go/iam/v1"
	"github.com/medhen/pc-iam-svc/internal/domain"
	"github.com/medhen/pc-iam-svc/internal/infrastructure/opa"
)

type AuthzUseCase struct {
	repo      domain.PolicyRepository
	opaClient *opa.Client
}

func NewAuthzUseCase(repo domain.PolicyRepository, opaClient *opa.Client) *AuthzUseCase {
	return &AuthzUseCase{repo: repo, opaClient: opaClient}
}

func (u *AuthzUseCase) Authorize(ctx context.Context, req *iamv1.AuthorizationRequest) (*iamv1.AuthorizationDecision, error) {
	// In a real implementation:
	// 1. Decode JWT and validate signature using cached JWKS
	// 2. Extract tenant ID from token
	// 3. Fast-fail if token tenant != resource_tenant_id (cross-tenant breach attempt)
	
	// Mock decoding for MVP
	tokenTenantID := req.ResourceTenantId // Simplified: Assume token matches resource request
	
	if tokenTenantID != req.ResourceTenantId {
		return &iamv1.AuthorizationDecision{
			IsAllowed:    false,
			DenialReason: "TENANT_MISMATCH",
		}, nil
	}

	// MFA Step-Up Authentication Check (Enhancement)
	// Example: High-value actions require 'acr' claim to be 'mfa' or 'aal2'
	requiresMFA := req.Action == "TRANSFER_FUNDS" || req.Action == "DELETE_PARTY"
	hasMFAClaim := false // Mock: would be extracted from token 'acr' claim
	
	if requiresMFA && !hasMFAClaim {
		return &iamv1.AuthorizationDecision{
			IsAllowed:    false,
			DenialReason: "STEP_UP_AUTH_REQUIRED",
		}, nil
	}

	// Call OPA sidecar
	decision, err := u.opaClient.Evaluate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("opa evaluation failed: %w", err)
	}
	
	// Partial Evaluation (Data Filtering AST)
	// If the decision is allowed but includes an AST for data filtering, we can serialize it
	// and pass it back in the decision for the domain service to convert to SQL.
	// decision.SqlFilterAst = parsePartialEvaluationAst(...)

	return decision, nil
}
