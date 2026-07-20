// Package domain is the claims bounded context: first-notice-of-loss (FNOL),
// reserving and fast-track settlement. Money is platform/money; GPS coordinates
// are float64 (they are not monetary).
package domain

import (
	"errors"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

var (
	ErrNegativeAmount    = errors.New("claims: amount cannot be negative")
	ErrAlreadySettled    = errors.New("claims: claim already settled")
	ErrAuthorityExceeded = errors.New("claims: settlement exceeds fast-track authority (refer)")
)

// Status is the claim lifecycle state.
type Status string

const (
	StatusFiled    Status = "FILED"
	StatusSettled  Status = "SETTLED"
	StatusRejected Status = "REJECTED"
)

// Claim is the aggregate root: a loss reported against a policy.
type Claim struct {
	ID            string
	TenantID      string
	PolicyID      string
	PartyID       string
	Status        Status
	Description   string
	Latitude      float64
	Longitude     float64
	Reserve       money.Amount
	SettledAmount money.Amount
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Version       int
}

// NewClaim files a FNOL with an initial indemnity reserve.
func NewClaim(tenantID, policyID, partyID, description string, lat, lng float64, reserve money.Amount) (*Claim, error) {
	if reserve.IsNegative() {
		return nil, ErrNegativeAmount
	}
	now := time.Now().UTC()
	return &Claim{
		ID: ids.New(), TenantID: tenantID, PolicyID: policyID, PartyID: partyID,
		Status: StatusFiled, Description: description, Latitude: lat, Longitude: lng,
		Reserve: reserve, SettledAmount: money.Zero(),
		CreatedAt: now, UpdatedAt: now, Version: 1,
	}, nil
}

// Settle pays and closes the claim if the amount is within fast-track authority.
// Amounts above authority are referred (ErrAuthorityExceeded).
func (c *Claim) Settle(amount, authority money.Amount) error {
	if c.Status == StatusSettled {
		return ErrAlreadySettled
	}
	if amount.IsNegative() {
		return ErrNegativeAmount
	}
	if !authority.IsZero() && amount.Cmp(authority) > 0 {
		return ErrAuthorityExceeded
	}
	c.SettledAmount = amount
	c.Status = StatusSettled
	c.touch()
	return nil
}

func (c *Claim) touch() {
	c.UpdatedAt = time.Now().UTC()
	c.Version++
}
