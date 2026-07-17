package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PolicySummary struct {
	ID           uuid.UUID `json:"id"`
	PolicyNumber string    `json:"policy_number"`
	PartyID      uuid.UUID `json:"party_id"`
	Status       string    `json:"status"`
}

type SearchPoliciesQuery struct {
	PartyID *uuid.UUID
	Status  *string
}

type SearchPoliciesHandler struct {
	pool *pgxpool.Pool
}

func NewSearchPoliciesHandler(pool *pgxpool.Pool) *SearchPoliciesHandler {
	return &SearchPoliciesHandler{pool: pool}
}

func (h *SearchPoliciesHandler) Handle(ctx context.Context, q SearchPoliciesQuery) ([]PolicySummary, error) {
	// Base query
	sql := `SELECT id, policy_number, party_id, status FROM policies WHERE 1=1`
	var args []interface{}
	argIdx := 1

	if q.PartyID != nil {
		sql += ` AND party_id = $` + string(rune(argIdx+'0'))
		args = append(args, *q.PartyID)
		argIdx++
	}

	if q.Status != nil {
		sql += ` AND status = $` + string(rune(argIdx+'0'))
		args = append(args, *q.Status)
		argIdx++
	}

	rows, err := h.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PolicySummary
	for rows.Next() {
		var s PolicySummary
		if err := rows.Scan(&s.ID, &s.PolicyNumber, &s.PartyID, &s.Status); err != nil {
			return nil, err
		}
		results = append(results, s)
	}

	return results, nil
}
