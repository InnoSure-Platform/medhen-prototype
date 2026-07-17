package grpc

import (
	"context"

	iamv1 "github.com/medhen/pc-contracts/gen/go/iam/v1"
	"github.com/medhen/pc-iam-svc/internal/application"
)

type Handler struct {
	iamv1.UnimplementedPolicyEnforcementServiceServer
	authzUseCase *application.AuthzUseCase
}

func NewHandler(authzUseCase *application.AuthzUseCase) *Handler {
	return &Handler{authzUseCase: authzUseCase}
}

func (h *Handler) AuthorizeAction(ctx context.Context, req *iamv1.AuthorizationRequest) (*iamv1.AuthorizationDecision, error) {
	return h.authzUseCase.Authorize(ctx, req)
}

func (h *Handler) IntrospectToken(ctx context.Context, req *iamv1.TokenString) (*iamv1.TokenClaims, error) {
	// MOCK offline JWT introspection logic using cached JWKS
	return &iamv1.TokenClaims{
		UserId:   "mock-user",
		TenantId: "mock-tenant",
		Roles:    []string{"mock-role"},
	}, nil
}
