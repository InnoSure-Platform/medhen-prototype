package domain

import "errors"

var (
	ErrInvalidRiskPayload     = errors.New("invalid risk payload")
	ErrReferralNotOpen        = errors.New("referral is not in OPEN state")
	ErrReferralAlreadyClosed  = errors.New("referral is already closed")
	ErrInsufficientAuthority  = errors.New("insufficient authority for decision")
	ErrInvalidDecision        = errors.New("invalid decision type")
	ErrFacultativeRequired    = errors.New("facultative reinsurance clearance required")
	ErrDisclosureMissing      = errors.New("required disclosures not acknowledged")
	ErrInvalidTransition      = errors.New("invalid state transition")
)
