// Package domain is the policy bounded context: quotes and the policies they
// bind into. Money is carried as platform/money (no float64), fixing the
// pre-refactor float premium in the policy aggregate.
package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

var (
	ErrQuoteNotBindable = errors.New("policy: quote is not in a bindable state")
	ErrDeclined         = errors.New("policy: underwriting declined the risk")
	ErrReferred         = errors.New("policy: underwriting referred the risk (manual review)")
	ErrNotInForce       = errors.New("policy: not in force")
	ErrAlreadyCancelled = errors.New("policy: already cancelled")
)

// QuoteStatus is the quote lifecycle state.
type QuoteStatus string

const (
	QuoteQuoted  QuoteStatus = "QUOTED"
	QuoteBound   QuoteStatus = "BOUND"
	QuoteExpired QuoteStatus = "EXPIRED"
)

// Quote is a priced offer for a party.
type Quote struct {
	ID             string
	TenantID       string
	PartyID        string
	ProductCode    string
	Coverages      []string
	RiskDimensions map[string]string
	NetPremium     money.Amount
	TotalTaxes     money.Amount
	GrossPremium   money.Amount
	CalculationID  string
	Status         QuoteStatus
	Version        int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewQuote creates a QUOTED offer.
func NewQuote(tenantID, partyID, productCode string, coverages []string, dims map[string]string,
	net, taxes, gross money.Amount, calcID string) *Quote {
	now := time.Now().UTC()
	return &Quote{
		ID: ids.New(), TenantID: tenantID, PartyID: partyID, ProductCode: productCode,
		Coverages: coverages, RiskDimensions: dims,
		NetPremium: net, TotalTaxes: taxes, GrossPremium: gross, CalculationID: calcID,
		Status: QuoteQuoted, Version: 1, CreatedAt: now, UpdatedAt: now,
	}
}

// Bind transitions a quote to BOUND. Only QUOTED quotes may bind.
func (q *Quote) Bind() error {
	if q.Status != QuoteQuoted {
		return ErrQuoteNotBindable
	}
	q.Status = QuoteBound
	q.UpdatedAt = time.Now().UTC()
	q.Version++
	return nil
}

// PolicyStatus is the policy lifecycle state.
type PolicyStatus string

const (
	StatusIssued    PolicyStatus = "ISSUED"
	StatusCancelled PolicyStatus = "CANCELLED"
)

// Policy is an in-force contract bound from a quote.
type Policy struct {
	ID            string
	PolicyNumber  string
	TenantID      string
	QuoteID       string
	PartyID       string
	ProductCode   string
	GrossPremium  money.Amount
	Status        PolicyStatus
	EffectiveFrom time.Time
	EffectiveTo   time.Time
	IssuedAt      time.Time
	Version       int
	// Servicing metadata.
	PriorPolicyID string     // set on a renewal successor
	CancelReason  string     // set on cancellation
	CancelledAt   *time.Time // set on cancellation
}

// NewPolicy issues a one-year policy from a bound quote.
func NewPolicy(policyNumber, tenantID, quoteID, partyID, productCode string, gross money.Amount, effectiveFrom time.Time) *Policy {
	return &Policy{
		ID: ids.New(), PolicyNumber: policyNumber, TenantID: tenantID, QuoteID: quoteID,
		PartyID: partyID, ProductCode: productCode, GrossPremium: gross,
		Status: StatusIssued, EffectiveFrom: effectiveFrom, EffectiveTo: effectiveFrom.AddDate(1, 0, 0),
		IssuedAt: time.Now().UTC(), Version: 1,
	}
}

// Endorse adjusts an in-force policy's gross premium by a (signed) delta,
// returning the new gross. Only ISSUED policies can be endorsed.
func (p *Policy) Endorse(premiumDelta money.Amount) error {
	if p.Status != StatusIssued {
		return ErrNotInForce
	}
	p.GrossPremium = p.GrossPremium.Add(premiumDelta)
	p.Version++
	return nil
}

// Cancel marks the policy CANCELLED with a reason. Returns the pro-rata unearned
// premium (a refund estimate) based on the remaining term at the cancel date.
func (p *Policy) Cancel(reason string, at time.Time) (money.Amount, error) {
	if p.Status == StatusCancelled {
		return money.Zero(), ErrAlreadyCancelled
	}
	refund := p.unearnedPremium(at)
	p.Status = StatusCancelled
	p.CancelReason = reason
	p.CancelledAt = &at
	p.Version++
	return refund, nil
}

// unearnedPremium is the pro-rata premium for the unexpired portion of the term.
func (p *Policy) unearnedPremium(at time.Time) money.Amount {
	total := p.EffectiveTo.Sub(p.EffectiveFrom).Hours()
	if total <= 0 {
		return money.Zero()
	}
	remaining := p.EffectiveTo.Sub(at).Hours()
	if remaining <= 0 {
		return money.Zero()
	}
	if remaining > total {
		remaining = total
	}
	return p.GrossPremium.Mul(decimal.NewFromFloat(remaining / total)).RoundCurrency()
}

// Renew builds a successor policy for the next term (same premium), linked back
// to this policy via PriorPolicyID. The caller assigns the new policy number.
func (p *Policy) Renew(policyNumber string) *Policy {
	next := NewPolicy(policyNumber, p.TenantID, p.QuoteID, p.PartyID, p.ProductCode, p.GrossPremium, p.EffectiveTo)
	next.PriorPolicyID = p.ID
	return next
}
