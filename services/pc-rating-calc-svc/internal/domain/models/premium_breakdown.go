package models

import (
	"github.com/shopspring/decimal"
)

// AuditStep records a single trace in the calculation pipeline
type AuditStep struct {
	StepOrder    int32  `json:"step_order"`
	Operation    string `json:"operation"`
	ValueApplied string `json:"value_applied"`
	TableVersion string `json:"table_version"`
	TraceID      string `json:"trace_id"`
	SpanID       string `json:"span_id"`
}

// CoveragePremium isolates the premium specific to one coverage
type CoveragePremium struct {
	CoverageCode string          `json:"coverage_code"`
	BasePremium  decimal.Decimal `json:"base_premium"`
	NetPremium   decimal.Decimal `json:"net_premium"`
}

// PremiumBreakdown is the immutable result of the RatingEngine
type PremiumBreakdown struct {
	CalculationID string          `json:"calculation_id"`
	NetPremium    decimal.Decimal `json:"net_premium"`
	TotalTaxes    decimal.Decimal `json:"total_taxes"`
	GrossPremium  decimal.Decimal `json:"gross_premium"`
	
	CoverageBreakdowns []CoveragePremium `json:"coverage_breakdowns"`
	TraceLog           []AuditStep       `json:"trace_log"`
}

// AddTrace appends an audit step securely
func (p *PremiumBreakdown) AddTrace(op, val, version, traceID, spanID string) {
	step := AuditStep{
		StepOrder:    int32(len(p.TraceLog) + 1),
		Operation:    op,
		ValueApplied: val,
		TableVersion: version,
		TraceID:      traceID,
		SpanID:       spanID,
	}
	p.TraceLog = append(p.TraceLog, step)
}
