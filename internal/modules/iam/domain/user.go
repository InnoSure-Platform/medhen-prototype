// Package domain is the IAM bounded context: application users and their role
// assignments within a tenant. Token verification itself lives in the platform
// auth kernel; this module owns application-level identity/authorization data.
package domain

import (
	"errors"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
)

var (
	ErrSubjectRequired = errors.New("iam: subject is required")
	ErrNoRoles         = errors.New("iam: at least one role is required")
)

// User is an application user mapped to an identity-provider subject.
type User struct {
	ID        string
	TenantID  string
	Subject   string // IdP subject (sub claim)
	Email     string
	FullName  string
	Roles     []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser validates and constructs a user.
func NewUser(tenantID, subject, email, fullName string, roles []string) (*User, error) {
	if subject == "" {
		return nil, ErrSubjectRequired
	}
	if len(roles) == 0 {
		return nil, ErrNoRoles
	}
	now := time.Now().UTC()
	return &User{
		ID: ids.New(), TenantID: tenantID, Subject: subject, Email: email,
		FullName: fullName, Roles: roles, CreatedAt: now, UpdatedAt: now,
	}, nil
}

// HasRole reports whether the user holds a role.
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}
