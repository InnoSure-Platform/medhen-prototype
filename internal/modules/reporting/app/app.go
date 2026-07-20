// Package app holds the reporting use cases: an event-sourced KPI read model
// (CQRS projection) computing real loss and combined ratios — replacing the
// pre-refactor hardcoded/dummy KPI values.
package app

import "context"

// assumedExpenseRatio is applied on top of the (real, data-derived) loss ratio to
// approximate a combined ratio until acquisition/admin expense feeds exist.
const assumedExpenseRatio = 0.30

// KPIRecord is the raw projected aggregate for a tenant.
type KPIRecord struct {
	TenantID            string
	PremiumWrittenMinor int64
	ClaimsPaidMinor     int64
	PolicyCount         int64
	ClaimCount          int64
}

// KPIView is the computed KPI response.
type KPIView struct {
	TenantID            string  `json:"tenant_id"`
	PremiumWrittenMinor int64   `json:"premium_written_minor"`
	ClaimsPaidMinor     int64   `json:"claims_paid_minor"`
	PolicyCount         int64   `json:"policy_count"`
	ClaimCount          int64   `json:"claim_count"`
	LossRatio           float64 `json:"loss_ratio"`
	AssumedExpenseRatio float64 `json:"assumed_expense_ratio"`
	CombinedRatio       float64 `json:"combined_ratio"`
}

// Repository maintains the KPI projection.
type Repository interface {
	AddPolicy(ctx context.Context, tenantID string, grossMinor int64) error
	AddClaim(ctx context.Context, tenantID string, amountMinor int64) error
	Get(ctx context.Context, tenantID string) (KPIRecord, error)
}

// Service implements reporting use cases.
type Service struct{ repo Repository }

// NewService builds the service.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// RecordPolicy folds an issued policy's premium into the projection.
func (s *Service) RecordPolicy(ctx context.Context, tenantID string, grossMinor int64) error {
	return s.repo.AddPolicy(ctx, tenantID, grossMinor)
}

// RecordClaim folds a settled claim's amount into the projection.
func (s *Service) RecordClaim(ctx context.Context, tenantID string, amountMinor int64) error {
	return s.repo.AddClaim(ctx, tenantID, amountMinor)
}

// KPIs returns the computed KPI view for a tenant. Loss ratio is derived from the
// projected figures (0 when no premium has been written).
func (s *Service) KPIs(ctx context.Context, tenantID string) (KPIView, error) {
	r, err := s.repo.Get(ctx, tenantID)
	if err != nil {
		return KPIView{}, err
	}
	var lossRatio float64
	if r.PremiumWrittenMinor > 0 {
		lossRatio = float64(r.ClaimsPaidMinor) / float64(r.PremiumWrittenMinor)
	}
	return KPIView{
		TenantID:            tenantID,
		PremiumWrittenMinor: r.PremiumWrittenMinor,
		ClaimsPaidMinor:     r.ClaimsPaidMinor,
		PolicyCount:         r.PolicyCount,
		ClaimCount:          r.ClaimCount,
		LossRatio:           lossRatio,
		AssumedExpenseRatio: assumedExpenseRatio,
		CombinedRatio:       lossRatio + assumedExpenseRatio,
	}, nil
}
