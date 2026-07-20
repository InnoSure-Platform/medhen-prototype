package domain_test

import (
	"testing"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/policy/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

func newQuote() *domain.Quote {
	return domain.NewQuote("eic", "party-1", "MOT", []string{"OD"}, map[string]string{"age_band": "adult"},
		money.FromInt(2000), money.FromInt(335), money.FromInt(2335), "calc-1")
}

func TestNewQuote(t *testing.T) {
	q := newQuote()
	if q.Status != domain.QuoteQuoted || q.ID == "" || q.GrossPremium.Minor() != 233500 {
		t.Fatalf("unexpected quote: %+v", q)
	}
}

func TestQuoteBind(t *testing.T) {
	q := newQuote()
	if err := q.Bind(); err != nil {
		t.Fatalf("bind quoted: %v", err)
	}
	if q.Status != domain.QuoteBound || q.Version != 2 {
		t.Fatalf("unexpected after bind: %+v", q)
	}
	// Second bind is rejected (not bindable).
	if err := q.Bind(); err != domain.ErrQuoteNotBindable {
		t.Fatalf("re-bind should fail, got %v", err)
	}
}

func TestQuoteBind_Expired(t *testing.T) {
	q := newQuote()
	q.Status = domain.QuoteExpired
	if err := q.Bind(); err != domain.ErrQuoteNotBindable {
		t.Fatalf("expired bind should fail, got %v", err)
	}
}

func TestNewPolicy(t *testing.T) {
	eff := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	p := domain.NewPolicy("EIC/MOT/2026/000001", "eic", "q1", "party-1", "MOT", money.FromInt(2335), eff)
	if p.Status != domain.StatusIssued || p.PolicyNumber != "EIC/MOT/2026/000001" {
		t.Fatalf("unexpected policy: %+v", p)
	}
	if !p.EffectiveTo.Equal(eff.AddDate(1, 0, 0)) {
		t.Fatalf("expected 1-year term, got %s → %s", p.EffectiveFrom, p.EffectiveTo)
	}
}
