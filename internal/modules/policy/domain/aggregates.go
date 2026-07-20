// Package domain is the policy bounded context: quotes and the policies they
// bind into. Money is carried as platform/money (no float64), fixing the
// pre-refactor float premium in the policy aggregate.
package domain

import (
	"errors"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

var (
	ErrQuoteNotBindable = errors.New("policy: quote is not in a bindable state")
	ErrDeclined         = errors.New("policy: underwriting declined the risk")
	ErrReferred         = errors.New("policy: underwriting referred the risk (manual review)")
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
