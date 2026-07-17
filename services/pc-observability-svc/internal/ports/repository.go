package ports

import (
	"context"

	"medhen.com/pc-observability-svc/internal/domain"
)

// SLORepository defines the driven port for SLO persistence.
type SLORepository interface {
	Save(ctx context.Context, slo *domain.SLO) error
	FindByID(ctx context.Context, id string) (*domain.SLO, error)
}

// TenantRepository defines the driven port for Tenant persistence.
type TenantRepository interface {
	Save(ctx context.Context, tenant *domain.Tenant) error
}

// MimirClient defines the driven port for pushing configuration to the Data Plane.
type MimirClient interface {
	PushRules(ctx context.Context, slo *domain.SLO) error
}
