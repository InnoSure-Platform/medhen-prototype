package domain

import (
	"time"
)

// PolicyFact represents a highly denormalized row in the OLAP store (ClickHouse)
// Optimized for sub-10ms aggregation of GWP, NWP, etc.
type PolicyFact struct {
	PolicyID            string    `ch:"policy_id"`
	PolicyNumber        string    `ch:"policy_number"`
	TenantID            string    `ch:"tenant_id"`
	LOB                 string    `ch:"lob"`
	ProductCode         string    `ch:"product_code"`
	BranchCode          string    `ch:"branch_code"`
	Status              string    `ch:"status"`
	EffectiveDate       time.Time `ch:"effective_date"`
	GrossWrittenPremium float64   `ch:"gross_written_premium"`
	NetWrittenPremium   float64   `ch:"net_written_premium"`
	LastEventTimestamp  time.Time `ch:"last_event_timestamp"`
	Sign                int8      `ch:"sign"` // Used for ClickHouse CollapsingMergeTree
}

// KPISummary represents the aggregated dashboard metrics
type KPISummary struct {
	TotalGWP      float64 `json:"total_gwp"`
	TotalNWP      float64 `json:"total_nwp"`
	PolicyCount   int64   `json:"policy_count"`
	LossRatio     float64 `json:"loss_ratio"`
	CombinedRatio float64 `json:"combined_ratio"`
}
