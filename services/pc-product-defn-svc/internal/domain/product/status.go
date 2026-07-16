package product

import (
	"errors"
)

// Status represents the lifecycle state of a Product.
type Status string

const (
	StatusDraft     Status = "DRAFT"
	StatusReview    Status = "REVIEW"
	StatusApproved  Status = "APPROVED"
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
	StatusRetired   Status = "RETIRED"
)

var (
	ErrInvalidStateTransition  = errors.New("invalid state transition")
	ErrProductNotInDraft       = errors.New("product must be in DRAFT state to be modified")
	ErrMissingFairValue        = errors.New("transition to REVIEW blocked: missing Fair Value Assessment")
)

// CanTransitionTo checks if the current status can transition to the new status.
func (s Status) CanTransitionTo(next Status) bool {
	switch s {
	case StatusDraft:
		return next == StatusReview || next == StatusRetired
	case StatusReview:
		return next == StatusApproved || next == StatusDraft // Can be rejected back to draft
	case StatusApproved:
		return next == StatusActive || next == StatusDraft
	case StatusActive:
		return next == StatusSuspended || next == StatusRetired
	case StatusSuspended:
		return next == StatusActive || next == StatusRetired
	case StatusRetired:
		return false // Terminal state
	default:
		return false
	}
}
