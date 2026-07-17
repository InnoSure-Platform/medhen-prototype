package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidTenantName = errors.New("tenant name cannot be empty")
)

// TenantStatus represents the current state of a tenant provisioning.
type TenantStatus string

const (
	TenantStatusPending TenantStatus = "PENDING"
	TenantStatusActive  TenantStatus = "ACTIVE"
)

// Tenant is the aggregate root for a logical isolation boundary.
type Tenant struct {
	ID         string
	Name       string
	Status     TenantStatus
	APIKeyHash string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewTenant creates a new Tenant aggregate.
func NewTenant(id, name, apiKeyHash string) (*Tenant, error) {
	if name == "" {
		return nil, ErrInvalidTenantName
	}

	return &Tenant{
		ID:         id,
		Name:       name,
		Status:     TenantStatusPending,
		APIKeyHash: apiKeyHash,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}, nil
}

// MarkActive marks the tenant as successfully provisioned in the Data Plane.
func (t *Tenant) MarkActive() {
	t.Status = TenantStatusActive
	t.UpdatedAt = time.Now().UTC()
}
