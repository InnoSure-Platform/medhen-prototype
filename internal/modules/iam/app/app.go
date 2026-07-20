// Package app holds the IAM use cases.
package app

import (
	"context"
	"errors"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam/domain"
)

// ErrNotFound is returned when a user does not exist.
var ErrNotFound = errors.New("iam: user not found")

// Repository persists application users.
type Repository interface {
	Save(ctx context.Context, u *domain.User) error
	Get(ctx context.Context, tenantID, id string) (*domain.User, error)
	GetBySubject(ctx context.Context, tenantID, subject string) (*domain.User, error)
}

// Service implements IAM use cases.
type Service struct{ repo Repository }

// NewService builds the service.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// RegisterInput is the command payload for provisioning a user.
type RegisterInput struct {
	TenantID string   `json:"tenant_id"`
	Subject  string   `json:"subject"`
	Email    string   `json:"email"`
	FullName string   `json:"full_name"`
	Roles    []string `json:"roles"`
}

// Register provisions an application user.
func (s *Service) Register(ctx context.Context, in RegisterInput) (*domain.User, error) {
	u, err := domain.NewUser(in.TenantID, in.Subject, in.Email, in.FullName, in.Roles)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Get loads a user by id.
func (s *Service) Get(ctx context.Context, tenantID, id string) (*domain.User, error) {
	return s.repo.Get(ctx, tenantID, id)
}

// RolesForSubject resolves a user's roles by subject (implements ports.Reader).
func (s *Service) RolesForSubject(ctx context.Context, tenantID, subject string) ([]string, error) {
	u, err := s.repo.GetBySubject(ctx, tenantID, subject)
	if err != nil {
		return nil, err
	}
	return u.Roles, nil
}
