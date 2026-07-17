package aggregate

import (
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-underwriting-svc/internal/domain"
	"github.com/medhen/pc-underwriting-svc/internal/domain/valueobject"
)

type ReferralStatus string

const (
	ReferralStatusOpen      ReferralStatus = "OPEN"
	ReferralStatusInReview  ReferralStatus = "IN_REVIEW"
	ReferralStatusEscalated ReferralStatus = "ESCALATED"
	ReferralStatusApproved  ReferralStatus = "APPROVED"
	ReferralStatusDeclined  ReferralStatus = "DECLINED"
	ReferralStatusInvalid   ReferralStatus = "INVALIDATED"
)

// Referral manages the human-in-the-loop workflow for a specific assessment.
type Referral struct {
	ID                     uuid.UUID
	AssessmentID           uuid.UUID
	TenantID               string
	Status                 ReferralStatus
	RequiredAuthorityLevel string
	AssignedTo             *string
	Decision               *valueobject.DecisionType
	Conditions             []valueobject.Condition
	DisclosuresAcked       []string
	FacultativeRequired    bool
	SLADeadline            time.Time
	Version                int
	CreatedAt              time.Time
	ResolvedAt             *time.Time
}

// NewReferral creates a new referral from an assessment.
func NewReferral(tenantID string, assessmentID uuid.UUID, authorityLevel string, slaHours int, isFacultative bool) *Referral {
	return &Referral{
		ID:                     uuid.New(),
		AssessmentID:           assessmentID,
		TenantID:               tenantID,
		Status:                 ReferralStatusOpen,
		RequiredAuthorityLevel: authorityLevel,
		FacultativeRequired:    isFacultative,
		SLADeadline:            time.Now().UTC().Add(time.Duration(slaHours) * time.Hour),
		Version:                1,
		CreatedAt:              time.Now().UTC(),
	}
}

// Assign assigns the referral to an underwriter queue.
func (r *Referral) Assign(underwriterID string) error {
	if r.Status != ReferralStatusOpen && r.Status != ReferralStatusEscalated {
		return domain.ErrReferralNotOpen
	}
	r.AssignedTo = &underwriterID
	r.Status = ReferralStatusInReview
	r.Version++
	return nil
}

// Decide processes an underwriter's decision.
func (r *Referral) Decide(decision valueobject.DecisionType, conditions []valueobject.Condition, disclosures []string, actorAuthority *AuthorityLevel, premium, tsi float64, facultativeCleared bool) error {
	if r.Status != ReferralStatusInReview {
		return domain.ErrReferralNotOpen
	}

	// Facultative check
	if r.FacultativeRequired && !facultativeCleared && decision != valueobject.DecisionDecline && decision != valueobject.DecisionReferHigher {
		return domain.ErrFacultativeRequired
	}

	// If approving, verify authority limits
	if decision == valueobject.DecisionApprove || decision == valueobject.DecisionApproveWithConditions {
		if actorAuthority == nil || !actorAuthority.CanApprove(premium, tsi) {
			return domain.ErrInsufficientAuthority
		}
	}

	r.Decision = &decision
	r.Conditions = conditions
	r.DisclosuresAcked = disclosures

	now := time.Now().UTC()
	r.ResolvedAt = &now

	switch decision {
	case valueobject.DecisionApprove, valueobject.DecisionApproveWithConditions:
		r.Status = ReferralStatusApproved
	case valueobject.DecisionDecline:
		r.Status = ReferralStatusDeclined
	case valueobject.DecisionReferHigher:
		r.Status = ReferralStatusEscalated
		if actorAuthority != nil {
			r.RequiredAuthorityLevel = actorAuthority.NextLevel()
		}
		// Refer higher resets resolution since it goes back to queue
		r.ResolvedAt = nil
		r.AssignedTo = nil
	default:
		return domain.ErrInvalidDecision
	}

	r.Version++
	return nil
}

// Invalidate invalidates a referral due to an upstream quote amendment.
func (r *Referral) Invalidate() error {
	if r.Status == ReferralStatusInvalid {
		return nil
	}
	r.Status = ReferralStatusInvalid
	now := time.Now().UTC()
	r.ResolvedAt = &now
	r.Version++
	return nil
}

// CheckSLA checks if the SLA is breached and automatically escalates.
func (r *Referral) CheckSLA() bool {
	if (r.Status == ReferralStatusOpen || r.Status == ReferralStatusInReview) && time.Now().UTC().After(r.SLADeadline) {
		r.Status = ReferralStatusEscalated
		// In a real system, we'd also boost the authority level or send a notification.
		r.AssignedTo = nil
		r.Version++
		return true
	}
	return false
}
