package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// TransactionState represents the lifecycle of a payment intent.
type TransactionState string

const (
	StateInitiated      TransactionState = "INITIATED"
	StatePendingPartner TransactionState = "PENDING_PARTNER"
	StateSuccess        TransactionState = "SUCCESS"
	StateFailed         TransactionState = "FAILED"
	StateExpired        TransactionState = "EXPIRED"
)

// IntegrationTransaction acts as the aggregate root for outbound calls.
type IntegrationTransaction struct {
	InternalReferenceID  uuid.UUID
	Provider             string
	TransactionType      string
	Amount               float64
	Currency             string
	State                TransactionState
	CircuitBreakerStatus string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

var (
	ErrInvalidStateTransition = errors.New("invalid state transition")
)

// NewIntegrationTransaction creates a new initiated transaction.
func NewIntegrationTransaction(refID uuid.UUID, provider, txnType string, amount float64, currency string) *IntegrationTransaction {
	now := time.Now()
	return &IntegrationTransaction{
		InternalReferenceID: refID,
		Provider:            provider,
		TransactionType:     txnType,
		Amount:              amount,
		Currency:            currency,
		State:               StateInitiated,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// MarkPending transitions the transaction to PENDING_PARTNER once sent.
func (t *IntegrationTransaction) MarkPending() error {
	if t.State != StateInitiated {
		return ErrInvalidStateTransition
	}
	t.State = StatePendingPartner
	t.UpdatedAt = time.Now()
	return nil
}

// MarkSuccess transitions to SUCCESS upon webhook callback.
func (t *IntegrationTransaction) MarkSuccess() error {
	if t.State != StatePendingPartner {
		return ErrInvalidStateTransition
	}
	t.State = StateSuccess
	t.UpdatedAt = time.Now()
	return nil
}

// MarkFailed transitions to FAILED upon webhook callback or synchronous error.
func (t *IntegrationTransaction) MarkFailed() error {
	if t.State != StatePendingPartner && t.State != StateInitiated {
		return ErrInvalidStateTransition
	}
	t.State = StateFailed
	t.UpdatedAt = time.Now()
	return nil
}
