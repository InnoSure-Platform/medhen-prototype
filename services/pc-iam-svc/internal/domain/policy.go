package domain

import "context"

type AccessPolicy struct {
	ID          string
	Name        string
	RegoContent string
	Version     int
	IsActive    bool
}

type PolicyRepository interface {
	SavePolicy(ctx context.Context, policy *AccessPolicy) error
	GetActivePolicies(ctx context.Context) ([]*AccessPolicy, error)
}
