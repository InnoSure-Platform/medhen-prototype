package domain

import (
	"errors"
	"time"
)

var (
	ErrPolicyInactive   = errors.New("policy was not active on date of loss")
	ErrCoverageExcluded = errors.New("loss type is not covered by policy")
	ErrClaimClosed      = errors.New("cannot mutate a closed claim")
)

type ClaimStatus string

const (
	FNOL               ClaimStatus = "FNOL"
	Registered         ClaimStatus = "REGISTERED"
	CoverageDenied     ClaimStatus = "COVERAGE_DENIED"
	Triaged            ClaimStatus = "TRIAGED"
	FastTrack          ClaimStatus = "FAST_TRACK"
	UnderInvestigation ClaimStatus = "UNDER_INVESTIGATION"
	Assessed           ClaimStatus = "ASSESSED"
	PendingApproval    ClaimStatus = "PENDING_APPROVAL"
	Approved           ClaimStatus = "APPROVED"
	Settled            ClaimStatus = "SETTLED"
	Closed             ClaimStatus = "CLOSED"
)

// LossDetails is a value object capturing the event
type LossDetails struct {
	DateOfLoss  time.Time
	LossType    string
	Description string
	Lat         float64
	Lon         float64
}

// Claim is the primary aggregate root orchestrating the lifecycle
type Claim struct {
	ID                     string
	TenantID               string
	ClaimNumber            string
	PolicyID               string
	Status                 ClaimStatus
	Version                int
	LossDetails            LossDetails
	FraudScore             int
	AssigneeID             *string
	IsVulnerableCustomer   bool
	SlaDeadline            time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func NewClaim(tenantID, policyID, lossType string, dateOfLoss time.Time, isVulnerable bool) *Claim {
	
	// FCA Consumer Duty: Tighten statutory SLA if vulnerable
	slaHours := 48
	if isVulnerable {
		slaHours = 24
	}

	return &Claim{
		TenantID:             tenantID,
		PolicyID:             policyID,
		Status:               FNOL,
		LossDetails: LossDetails{
			DateOfLoss: dateOfLoss,
			LossType:   lossType,
		},
		IsVulnerableCustomer: isVulnerable,
		SlaDeadline:          time.Now().Add(time.Duration(slaHours) * time.Hour),
		Version:              1,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

// ValidateCoverage transitions the claim based on the upstream policy response
func (c *Claim) ValidateCoverage(isActive bool, isCovered bool) error {
	if c.Status != FNOL {
		return errors.New("can only validate coverage from FNOL state")
	}

	if !isActive || !isCovered {
		c.Status = CoverageDenied
		return nil
	}

	c.Status = Registered
	return nil
}

// Triage assigns the claim or routes to fast-track
func (c *Claim) Triage(fraudScore int, stpEligible bool) {
	c.FraudScore = fraudScore
	if fraudScore > 80 {
		// Route to SIU
		c.Status = UnderInvestigation
		return
	}

	if stpEligible {
		c.Status = FastTrack
	} else {
		c.Status = Triaged
	}
}
