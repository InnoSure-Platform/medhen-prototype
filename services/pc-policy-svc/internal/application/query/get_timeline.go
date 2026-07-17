package query

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TimelineEntry struct {
	VersionSeq    int       `json:"version_seq"`
	EffectiveFrom time.Time `json:"effective_from"`
	EffectiveTo   time.Time `json:"effective_to,omitempty"`
	SystemFrom    time.Time `json:"system_from"`
	SystemTo      time.Time `json:"system_to,omitempty"`
	TotalPremium  float64   `json:"total_premium"`
	RiskPayload   []byte    `json:"risk_payload"`
}

type GetTimelineQuery struct {
	PolicyID uuid.UUID
}

type GetTimelineHandler struct {
	pool *pgxpool.Pool
}

func NewGetTimelineHandler(pool *pgxpool.Pool) *GetTimelineHandler {
	return &GetTimelineHandler{pool: pool}
}

func (h *GetTimelineHandler) Handle(ctx context.Context, q GetTimelineQuery) ([]TimelineEntry, error) {
	// Query to extract exact tstzrange boundaries
	sql := `
		SELECT 
			version_seq, 
			lower(effective_period), 
			upper(effective_period), 
			lower(system_period), 
			upper(system_period), 
			total_premium, 
			risk_payload
		FROM policy_versions
		WHERE policy_id = $1
		ORDER BY version_seq ASC, lower(system_period) ASC
	`

	rows, err := h.pool.Query(ctx, sql, q.PolicyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timeline []TimelineEntry

	for rows.Next() {
		var entry TimelineEntry
		var effTo, sysTo *time.Time

		err := rows.Scan(
			&entry.VersionSeq,
			&entry.EffectiveFrom,
			&effTo,
			&entry.SystemFrom,
			&sysTo,
			&entry.TotalPremium,
			&entry.RiskPayload,
		)
		if err != nil {
			return nil, err
		}

		if effTo != nil {
			entry.EffectiveTo = *effTo
		}
		if sysTo != nil {
			entry.SystemTo = *sysTo
		}

		timeline = append(timeline, entry)
	}

	return timeline, nil
}
