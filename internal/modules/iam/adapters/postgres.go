// Package adapters holds the IAM module's Postgres repository.
package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	iamapp "github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam/app"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/iam/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/database"
)

// Schema is the DDL for the IAM module's table.
const Schema = `
CREATE TABLE IF NOT EXISTS iam_users (
    id         TEXT PRIMARY KEY,
    tenant_id  TEXT NOT NULL,
    subject    TEXT NOT NULL,
    email      TEXT NOT NULL DEFAULT '',
    full_name  TEXT NOT NULL DEFAULT '',
    roles      JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (tenant_id, subject)
);
`

// UserRepository implements app.Repository.
type UserRepository struct{ db *database.DB }

// NewUserRepository builds the repository.
func NewUserRepository(db *database.DB) *UserRepository { return &UserRepository{db: db} }

var _ iamapp.Repository = (*UserRepository)(nil)

// Save upserts a user using the ambient connection.
func (r *UserRepository) Save(ctx context.Context, u *domain.User) error {
	roles, _ := json.Marshal(u.Roles)
	_, err := r.db.Conn(ctx).Exec(ctx,
		`INSERT INTO iam_users (id, tenant_id, subject, email, full_name, roles, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 ON CONFLICT (tenant_id, subject) DO UPDATE SET email=EXCLUDED.email,
		   full_name=EXCLUDED.full_name, roles=EXCLUDED.roles, updated_at=EXCLUDED.updated_at`,
		u.ID, u.TenantID, u.Subject, u.Email, u.FullName, roles, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("iam repo: save: %w", err)
	}
	return nil
}

func (r *UserRepository) scan(row pgx.Row) (*domain.User, error) {
	var u domain.User
	var roles []byte
	err := row.Scan(&u.ID, &u.TenantID, &u.Subject, &u.Email, &u.FullName, &roles, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, iamapp.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("iam repo: scan: %w", err)
	}
	_ = json.Unmarshal(roles, &u.Roles)
	return &u, nil
}

const cols = `id, tenant_id, subject, email, full_name, roles, created_at, updated_at`

// Get loads a user by id.
func (r *UserRepository) Get(ctx context.Context, tenantID, id string) (*domain.User, error) {
	return r.scan(r.db.Conn(ctx).QueryRow(ctx,
		`SELECT `+cols+` FROM iam_users WHERE tenant_id=$1 AND id=$2`, tenantID, id))
}

// GetBySubject loads a user by IdP subject.
func (r *UserRepository) GetBySubject(ctx context.Context, tenantID, subject string) (*domain.User, error) {
	return r.scan(r.db.Conn(ctx).QueryRow(ctx,
		`SELECT `+cols+` FROM iam_users WHERE tenant_id=$1 AND subject=$2`, tenantID, subject))
}
