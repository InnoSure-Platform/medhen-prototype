// Package ports is the published contract of the underwriting module.
package ports

import (
	"context"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

// Decision outcomes.
const (
	OutcomeAutoAccept = "AUTO_ACCEPT" // straight-through processing
	OutcomeRefer      = "REFER"       // needs human underwriter
	OutcomeDecline    = "DECLINE"
)

// DecisionRequest is the input to an underwriting decision.
type DecisionRequest struct {
	TenantID       string            `json:"tenant_id"`
	ProductCode    string            `json:"product_code"`
	GrossPremium   money.Amount      `json:"gross_premium"`
	RiskDimensions map[string]string `json:"risk_dimensions"`
}

// Decision is the underwriting outcome with human-readable reasons.
type Decision struct {
	Outcome string   `json:"outcome"`
	Reasons []string `json:"reasons"`
}

// Accepted reports whether the risk may bind straight through.
func (d Decision) Accepted() bool { return d.Outcome == OutcomeAutoAccept }

// Decider evaluates a risk for straight-through processing. Consumed in-process
// by the policy module during bind.
type Decider interface {
	Decide(ctx context.Context, req DecisionRequest) (Decision, error)
}
