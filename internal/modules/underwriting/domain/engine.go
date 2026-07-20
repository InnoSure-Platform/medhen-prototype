// Package domain is the underwriting bounded context: rules that decide whether
// a risk can bind straight-through, must be referred, or is declined. Pure domain.
package domain

import (
	"context"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/underwriting/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

// Rules configures the STP decision thresholds.
type Rules struct {
	// ReferAbove refers risks whose gross premium exceeds this amount to a human.
	ReferAbove money.Amount
	// MaxPriorClaims refers risks with more than this many prior claims.
	MaxPriorClaims int
}

// Engine applies underwriting rules. Stateless and concurrency-safe.
type Engine struct{ rules Rules }

// NewEngine builds the engine.
func NewEngine(rules Rules) *Engine { return &Engine{rules: rules} }

var _ ports.Decider = (*Engine)(nil)

// Decide evaluates the risk. Any triggered rule downgrades AUTO_ACCEPT to REFER;
// a blacklisted risk is declined outright.
func (e *Engine) Decide(_ context.Context, req ports.DecisionRequest) (ports.Decision, error) {
	if req.RiskDimensions["blacklisted"] == "true" {
		return ports.Decision{Outcome: ports.OutcomeDecline, Reasons: []string{"party is blacklisted"}}, nil
	}

	var reasons []string
	if !e.rules.ReferAbove.IsZero() && req.GrossPremium.Cmp(e.rules.ReferAbove) > 0 {
		reasons = append(reasons, "gross premium exceeds STP threshold")
	}
	if priorClaims(req.RiskDimensions) > e.rules.MaxPriorClaims {
		reasons = append(reasons, "prior claims exceed STP threshold")
	}

	if len(reasons) > 0 {
		return ports.Decision{Outcome: ports.OutcomeRefer, Reasons: reasons}, nil
	}
	return ports.Decision{Outcome: ports.OutcomeAutoAccept, Reasons: []string{"within STP rules"}}, nil
}

func priorClaims(dims map[string]string) int {
	switch dims["prior_claims"] {
	case "", "0":
		return 0
	case "1":
		return 1
	case "2":
		return 2
	default:
		return 99 // any other value is treated as "many"
	}
}
