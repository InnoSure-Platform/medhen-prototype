package domain

import (
	"context"
)

// SCIM 2.0 Core User Schema
type SCIMUser struct {
	Schemas  []string   `json:"schemas"`
	ID       string     `json:"id,omitempty"`
	UserName string     `json:"userName"`
	Name     SCIMName   `json:"name"`
	Emails   []SCIMEmail `json:"emails"`
	Active   bool       `json:"active"`
}

type SCIMName struct {
	Formatted  string `json:"formatted"`
	FamilyName string `json:"familyName"`
	GivenName  string `json:"givenName"`
}

type SCIMEmail struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary"`
}

type SCIMRepository interface {
	CreateUser(ctx context.Context, tenantID string, user *SCIMUser) (*SCIMUser, error)
	GetUser(ctx context.Context, tenantID string, id string) (*SCIMUser, error)
	UpdateUser(ctx context.Context, tenantID string, id string, user *SCIMUser) (*SCIMUser, error)
	DeleteUser(ctx context.Context, tenantID string, id string) error
}
