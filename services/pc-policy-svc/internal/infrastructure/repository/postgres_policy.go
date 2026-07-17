package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-policy-svc/internal/domain/policy"
)

type PostgresPolicyRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPolicyRepository(pool *pgxpool.Pool) *PostgresPolicyRepository {
	return &PostgresPolicyRepository{pool: pool}
}

func (r *PostgresPolicyRepository) Save(ctx context.Context, p *policy.Policy) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = r.SaveWithTx(ctx, tx, p)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresPolicyRepository) SaveWithTx(ctx context.Context, tx pgx.Tx, p *policy.Policy) error {
	// Upsert Policy aggregate root
	policyQuery := `
		INSERT INTO policies (id, tenant_id, policy_number, product_id, party_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status
	`
	_, err = tx.Exec(ctx, policyQuery, p.ID, p.TenantID, p.PolicyNumber, p.ProductID, p.PartyID, string(p.Status), p.CreatedAt)
	if err != nil {
		return err
	}

	// Insert new policy versions
	versionQuery := `
		INSERT INTO policy_versions (id, policy_id, version_seq, risk_payload, total_premium, effective_period, system_period)
		VALUES ($1, $2, $3, $4, $5, tstzrange($6, $7, '[)'), tstzrange($8, $9, '[)'))
		ON CONFLICT (id) DO NOTHING
	`

	for _, v := range p.Versions {
		payload, _ := json.Marshal(v.RiskPayload)
		_, err = tx.Exec(ctx, versionQuery, v.ID, v.PolicyID, v.VersionSeq, payload, v.TotalPremium, v.EffectiveFrom, v.EffectiveTo, v.SystemFrom, v.SystemTo)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresPolicyRepository) GetByID(ctx context.Context, id uuid.UUID) (*policy.Policy, error) {
	// Simple implementation for fetching the policy with its latest valid system_time version
	// A proper implementation would fetch all versions or the requested temporal slice
	
	query := `SELECT tenant_id, policy_number, product_id, party_id, status, created_at FROM policies WHERE id = $1`
	var p policy.Policy
	p.ID = id
	var status string

	err := r.pool.QueryRow(ctx, query, id).Scan(&p.TenantID, &p.PolicyNumber, &p.ProductID, &p.PartyID, &status, &p.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("policy not found")
		}
		return nil, err
	}
	p.Status = policy.Status(status)

	// Fetch current active versions
	vQuery := `
		SELECT id, version_seq, risk_payload, total_premium, lower(effective_period), upper(effective_period), lower(system_period), upper(system_period)
		FROM policy_versions 
		WHERE policy_id = $1 AND upper_inf(system_period)
		ORDER BY version_seq ASC
	`
	rows, err := r.pool.Query(ctx, vQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var v policy.PolicyVersion
		v.PolicyID = id
		var payload []byte
		
		err := rows.Scan(&v.ID, &v.VersionSeq, &payload, &v.TotalPremium, &v.EffectiveFrom, &v.EffectiveTo, &v.SystemFrom, &v.SystemTo)
		if err != nil {
			return nil, err
		}
		v.RiskPayload = payload
		p.Versions = append(p.Versions, v)
	}

	return &p, nil
}

func (r *PostgresPolicyRepository) GetVersionAt(ctx context.Context, policyID uuid.UUID, asOf time.Time) (*policy.PolicyVersion, error) {
	query := `
		SELECT id, version_seq, risk_payload, total_premium, lower(effective_period), upper(effective_period), lower(system_period), upper(system_period)
		FROM policy_versions 
		WHERE policy_id = $1 AND effective_period @> $2::timestamptz AND upper_inf(system_period)
	`
	var v policy.PolicyVersion
	v.PolicyID = policyID
	var payload []byte

	err := r.pool.QueryRow(ctx, query, policyID, asOf).Scan(&v.ID, &v.VersionSeq, &payload, &v.TotalPremium, &v.EffectiveFrom, &v.EffectiveTo, &v.SystemFrom, &v.SystemTo)
	if err != nil {
		return nil, err
	}
	v.RiskPayload = payload
	return &v, nil
}
