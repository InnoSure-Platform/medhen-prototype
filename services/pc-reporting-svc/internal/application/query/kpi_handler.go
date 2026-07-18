package query

import (
	"context"
	"fmt"

	"github.com/medhen/pc-reporting-svc/internal/domain"
	"github.com/medhen/pc-reporting-svc/internal/infrastructure/clickhouse"
)

type KPIHandler struct {
	repo *clickhouse.QueryRepository
}

func NewKPIHandler(repo *clickhouse.QueryRepository) *KPIHandler {
	return &KPIHandler{repo: repo}
}

func (h *KPIHandler) Handle(ctx context.Context, tenantID string, lob string) (*domain.KPISummary, error) {
	summary, err := h.repo.GetKPISummary(ctx, tenantID, lob)
	if err != nil {
		return nil, fmt.Errorf("query repository error: %w", err)
	}
	return summary, nil
}
