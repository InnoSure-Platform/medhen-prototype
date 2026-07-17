package domain

import (
	"time"
)

type SignatureState string

const (
	SigStatePending  SignatureState = "SIGNATURE_PENDING"
	SigStateSigned   SignatureState = "SIGNED"
	SigStateDeclined SignatureState = "SIGNATURE_DECLINED"
)

type SignatureRequest struct {
	ID         string
	DocumentID string
	TenantID   string
	Signatory  string
	State      SignatureState
	IPAddress  string
	UserAgent  string
	SignedAt   *time.Time
	CreatedAt  time.Time
}

func NewSignatureRequest(id, documentID, tenantID, signatory string) *SignatureRequest {
	return &SignatureRequest{
		ID:         id,
		DocumentID: documentID,
		TenantID:   tenantID,
		Signatory:  signatory,
		State:      SigStatePending,
		CreatedAt:  time.Now().UTC(),
	}
}

func (s *SignatureRequest) Sign(ipAddress, userAgent string) error {
	if s.State != SigStatePending {
		return ErrSignatureNotPending
	}
	now := time.Now().UTC()
	s.State = SigStateSigned
	s.IPAddress = ipAddress
	s.UserAgent = userAgent
	s.SignedAt = &now
	return nil
}

func (s *SignatureRequest) Decline() error {
	if s.State != SigStatePending {
		return ErrSignatureNotPending
	}
	s.State = SigStateDeclined
	return nil
}
