package aggregate

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING_APPROVAL"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
)

// ApprovalRequest orchestrates the Maker-Checker principle for sensitive operations.
type ApprovalRequest struct {
	ID             uuid.UUID
	TenantID       string
	OperationType  string // e.g., "MANUAL_ALLOCATION", "REFUND_AUTHORIZATION"
	TargetID       uuid.UUID
	Payload        []byte
	Status         ApprovalStatus
	MakerID        string // User ID of the initiator
	CheckerID      *string // User ID of the approver
	RejectionNotes *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewApprovalRequest(tenantID, opType string, targetID uuid.UUID, payload []byte, makerID string) *ApprovalRequest {
	return &ApprovalRequest{
		ID:            uuid.New(),
		TenantID:      tenantID,
		OperationType: opType,
		TargetID:      targetID,
		Payload:       payload,
		Status:        ApprovalStatusPending,
		MakerID:       makerID,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
}

func (a *ApprovalRequest) Approve(checkerID string) error {
	if a.Status != ApprovalStatusPending {
		return errors.New("only pending requests can be approved")
	}
	if a.MakerID == checkerID {
		return errors.New("maker and checker cannot be the same user")
	}

	a.Status = ApprovalStatusApproved
	a.CheckerID = &checkerID
	a.UpdatedAt = time.Now().UTC()
	return nil
}

func (a *ApprovalRequest) Reject(checkerID, notes string) error {
	if a.Status != ApprovalStatusPending {
		return errors.New("only pending requests can be rejected")
	}

	a.Status = ApprovalStatusRejected
	a.CheckerID = &checkerID
	a.RejectionNotes = &notes
	a.UpdatedAt = time.Now().UTC()
	return nil
}
