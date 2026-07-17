package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-iam-svc/internal/domain"
	"github.com/medhen/pc-iam-svc/internal/infrastructure/keycloak"
)

type TenantUseCase struct {
	repo     domain.TenantRepository
	kcClient *keycloak.Client
}

func NewTenantUseCase(repo domain.TenantRepository, kcClient *keycloak.Client) *TenantUseCase {
	return &TenantUseCase{repo: repo, kcClient: kcClient}
}

func (u *TenantUseCase) ProvisionTenant(ctx context.Context, name string) (*domain.Tenant, error) {
	tenantID := uuid.New().String()
	realmName := fmt.Sprintf("tenant-%s", tenantID)

	// 1. Provision Keycloak Realm
	if err := u.kcClient.CreateRealm(ctx, realmName); err != nil {
		return nil, fmt.Errorf("failed to create Keycloak realm: %w", err)
	}

	// 2. Persist to local DB
	tenant := &domain.Tenant{
		ID:              tenantID,
		Name:            name,
		KeycloakRealmID: realmName,
		Status:          "ACTIVE",
		CreatedAt:       time.Now(),
	}

	if err := u.repo.CreateTenant(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to save tenant to DB: %w", err)
	}

	return tenant, nil
}
