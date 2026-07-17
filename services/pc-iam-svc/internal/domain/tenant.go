package domain

import (
	"context"
	"time"
)

type Tenant struct {
	ID              string
	Name            string
	KeycloakRealmID string
	Status          string // ACTIVE, SUSPENDED
	CreatedAt       time.Time
}

type TenantRepository interface {
	CreateTenant(ctx context.Context, tenant *Tenant) error
	GetTenantByID(ctx context.Context, id string) (*Tenant, error)
	UpdateTenantStatus(ctx context.Context, id string, status string) error
}
