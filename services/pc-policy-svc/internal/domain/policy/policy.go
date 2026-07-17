package policy

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Status represents the lifecycle status of a policy
type Status string

const (
	StatusBound     Status = "BOUND"
	StatusActive    Status = "ACTIVE"
	StatusEndorsed  Status = "ENDORSED"
	StatusCancelled Status = "CANCELLED"
	StatusExpired   Status = "EXPIRED"
	StatusSuspended Status = "SUSPENDED"
)

var (
	ErrInvalidEndorsementDate = errors.New("endorsement date cannot precede policy inception")
	ErrPolicyAlreadyCancelled = errors.New("policy is already cancelled")
	ErrPolicyNotActive        = errors.New("policy is not active")
)

// Policy represents the formalized contract aggregate root.
type Policy struct {
	ID           uuid.UUID
	TenantID     string
	PolicyNumber string
	ProductID    uuid.UUID
	PartyID      uuid.UUID
	Status       Status
	CreatedAt    time.Time
	
	Versions     []PolicyVersion
}

// PolicyVersion represents a bi-temporal snapshot of a policy.
type PolicyVersion struct {
	ID             uuid.UUID
	PolicyID       uuid.UUID
	VersionSeq     int
	RiskPayload    []byte // JSONB
	TotalPremium   float64

	// Bi-temporal axes
	EffectiveFrom  time.Time
	EffectiveTo    *time.Time
	SystemFrom     time.Time
	SystemTo       *time.Time
}

// NewPolicy creates a new Policy from a bound quote.
func NewPolicy(tenantID, policyNumber string, productID, partyID uuid.UUID, initialPayload []byte, premium float64, effectiveFrom time.Time, effectiveTo time.Time) (*Policy, error) {
	now := time.Now()
	
	policyID := uuid.New()
	versionID := uuid.New()

	p := &Policy{
		ID:           policyID,
		TenantID:     tenantID,
		PolicyNumber: policyNumber,
		ProductID:    productID,
		PartyID:      partyID,
		Status:       StatusBound,
		CreatedAt:    now,
	}

	v := PolicyVersion{
		ID:            versionID,
		PolicyID:      policyID,
		VersionSeq:    1,
		RiskPayload:   initialPayload,
		TotalPremium:  premium,
		EffectiveFrom: effectiveFrom,
		EffectiveTo:   &effectiveTo,
		SystemFrom:    now,
		SystemTo:      nil, // Represents infinity in PostgreSQL
	}

	p.Versions = append(p.Versions, v)
	return p, nil
}

// Endorse creates a mid-term adjustment, creating a new bi-temporal version.
func (p *Policy) Endorse(effectiveDate time.Time, newPayload []byte, newPremium float64) (*PolicyVersion, error) {
	if p.Status == StatusCancelled || p.Status == StatusExpired {
		return nil, ErrPolicyNotActive
	}

	if len(p.Versions) == 0 {
		return nil, errors.New("corrupted policy state: no versions found")
	}

	// Simplistic approach for demo: get latest version
	latest := p.Versions[len(p.Versions)-1]

	if effectiveDate.Before(latest.EffectiveFrom) {
		return nil, ErrInvalidEndorsementDate
	}

	now := time.Now()

	// "Close" the current version's system time (In reality, the repository handles the tsrange updates, but we model the logical change here)
	// For pure bi-temporal, we create a new system record for the past, and a new effective record for the future.

	// Truncate the previous version's effective period to the start of this endorsement
	p.Versions[len(p.Versions)-1].EffectiveTo = &effectiveDate

	newVersion := PolicyVersion{
		ID:            uuid.New(),
		PolicyID:      p.ID,
		VersionSeq:    latest.VersionSeq + 1,
		RiskPayload:   newPayload,
		TotalPremium:  newPremium,
		EffectiveFrom: effectiveDate,
		EffectiveTo:   latest.EffectiveTo, // Inherit the original end date
		SystemFrom:    now,
		SystemTo:      nil,
	}

	p.Versions = append(p.Versions, newVersion)
	p.Status = StatusEndorsed

	return &newVersion, nil
}

// Cancel calculates a short-rate or pro-rata cancellation, returning a final refund amount.
func (p *Policy) Cancel(effectiveDate time.Time, reason string, isProRata bool) (float64, error) {
	if p.Status != StatusBound && p.Status != StatusActive && p.Status != StatusEndorsed && p.Status != StatusSuspended {
		return 0, ErrPolicyAlreadyCancelled
	}

	latest := p.Versions[len(p.Versions)-1]
	if effectiveDate.Before(latest.EffectiveFrom) {
		return 0, ErrInvalidEndorsementDate
	}

	now := time.Now()

	// Terminate the policy effective period
	newVersion := PolicyVersion{
		ID:            uuid.New(),
		PolicyID:      p.ID,
		VersionSeq:    latest.VersionSeq + 1,
		RiskPayload:   latest.RiskPayload,
		TotalPremium:  latest.TotalPremium,
		EffectiveFrom: effectiveDate,
		EffectiveTo:   &effectiveDate, // Effectively ends on this date
		SystemFrom:    now,
		SystemTo:      nil,
	}

	p.Versions = append(p.Versions, newVersion)
	p.Status = StatusCancelled

	// Simplistic refund logic for demo
	refund := 0.0
	if isProRata {
		refund = latest.TotalPremium * 0.10 // Simulate 10% refund
	} else {
		refund = latest.TotalPremium * 0.05 // Simulate 5% short-rate refund
	}

	return refund, nil
}

// Renew returns the payload necessary to create a new Draft Quote based on the current active policy state.
func (p *Policy) Renew() ([]byte, error) {
	if p.Status == StatusCancelled || p.Status == StatusExpired {
		return nil, ErrPolicyNotActive
	}

	latest := p.Versions[len(p.Versions)-1]
	
	return latest.RiskPayload, nil
}

