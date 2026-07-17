package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/medhen/pc-iam-svc/internal/domain"
	"github.com/medhen/pc-iam-svc/internal/infrastructure/keycloak"
)

type UserUseCase struct {
	repo     domain.UserRepository
	kcClient *keycloak.Client
}

func NewUserUseCase(repo domain.UserRepository, kcClient *keycloak.Client) *UserUseCase {
	return &UserUseCase{repo: repo, kcClient: kcClient}
}

func (u *UserUseCase) ProvisionUser(ctx context.Context, tenantID, partyID, username, email, initialPassword string) (*domain.UserIdentity, error) {
	// Need to get tenant to find realm name (simplified here assuming realm=tenant-<id>)
	realmName := fmt.Sprintf("tenant-%s", tenantID)

	// 1. Create in Keycloak
	kcUserID, err := u.kcClient.CreateUser(ctx, realmName, username, email, initialPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create Keycloak user: %w", err)
	}

	// 2. Map in local DB
	user := &domain.UserIdentity{
		ID:             uuid.New().String(),
		TenantID:       tenantID,
		KeycloakUserID: kcUserID,
		PartyID:        partyID,
		Status:         "ACTIVE",
	}

	if err := u.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user to DB: %w", err)
	}

	return user, nil
}
