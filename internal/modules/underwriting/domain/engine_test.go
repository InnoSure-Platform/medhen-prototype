package domain_test

import (
	"context"
	"testing"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/underwriting/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/underwriting/ports"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

func engine() *domain.Engine {
	return domain.NewEngine(domain.Rules{ReferAbove: money.FromInt(100000), MaxPriorClaims: 1})
}

func TestDecide_AutoAccept(t *testing.T) {
	d, _ := engine().Decide(context.Background(), ports.DecisionRequest{
		GrossPremium: money.FromInt(2680), RiskDimensions: map[string]string{"prior_claims": "0"},
	})
	if !d.Accepted() {
		t.Fatalf("expected AUTO_ACCEPT, got %+v", d)
	}
}

func TestDecide_ReferOnHighPremium(t *testing.T) {
	d, _ := engine().Decide(context.Background(), ports.DecisionRequest{
		GrossPremium: money.FromInt(250000),
	})
	if d.Outcome != ports.OutcomeRefer {
		t.Fatalf("expected REFER, got %+v", d)
	}
}

func TestDecide_ReferOnPriorClaims(t *testing.T) {
	d, _ := engine().Decide(context.Background(), ports.DecisionRequest{
		GrossPremium: money.FromInt(2680), RiskDimensions: map[string]string{"prior_claims": "2"},
	})
	if d.Outcome != ports.OutcomeRefer {
		t.Fatalf("expected REFER, got %+v", d)
	}
}

func TestDecide_DeclineBlacklisted(t *testing.T) {
	d, _ := engine().Decide(context.Background(), ports.DecisionRequest{
		RiskDimensions: map[string]string{"blacklisted": "true"},
	})
	if d.Outcome != ports.OutcomeDecline {
		t.Fatalf("expected DECLINE, got %+v", d)
	}
}
