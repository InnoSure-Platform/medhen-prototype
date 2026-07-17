package domain

import (
	"context"
)

type UserIdentity struct {
	ID             string
	TenantID       string
	KeycloakUserID string
	PartyID        string // Link to BC-MDH-01
	Status         string // ACTIVE, SUSPENDED
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *UserIdentity) error
	GetUserByID(ctx context.Context, id string) (*UserIdentity, error)
}
