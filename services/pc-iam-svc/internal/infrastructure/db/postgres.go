package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-iam-svc/internal/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(ctx context.Context, connString string) (*PostgresRepository, error) {
	poolCfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}
	
	poolCfg.MaxConns = 50
	poolCfg.MinConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &PostgresRepository{db: pool}, nil
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *PostgresRepository) Close() {
	r.db.Close()
}

// -- TenantRepository Implementation --

func (r *PostgresRepository) CreateTenant(ctx context.Context, tenant *domain.Tenant) error {
	query := `INSERT INTO tenants (id, name, keycloak_realm_id, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, tenant.ID, tenant.Name, tenant.KeycloakRealmID, tenant.Status, tenant.CreatedAt)
	return err
}

func (r *PostgresRepository) GetTenantByID(ctx context.Context, id string) (*domain.Tenant, error) {
	// MOCK
	return &domain.Tenant{}, nil
}

func (r *PostgresRepository) UpdateTenantStatus(ctx context.Context, id string, status string) error {
	// MOCK
	return nil
}

// -- UserRepository Implementation --

func (r *PostgresRepository) CreateUser(ctx context.Context, user *domain.UserIdentity) error {
	query := `INSERT INTO user_identities (id, tenant_id, keycloak_user_id, party_id, status) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, user.ID, user.TenantID, user.KeycloakUserID, user.PartyID, user.Status)
	return err
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id string) (*domain.UserIdentity, error) {
	// MOCK
	return &domain.UserIdentity{}, nil
}

// -- PolicyRepository Implementation --

func (r *PostgresRepository) SavePolicy(ctx context.Context, policy *domain.AccessPolicy) error {
	// MOCK
	return nil
}

func (r *PostgresRepository) GetActivePolicies(ctx context.Context) ([]*domain.AccessPolicy, error) {
	// MOCK
	return nil, nil
}
