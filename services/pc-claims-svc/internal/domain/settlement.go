package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrReserveExceeded = errors.New("settlement amount exceeds authorized indemnity reserve")
)

type SettlementStatus string

const (
	SettlementPending  SettlementStatus = "PENDING"
	SettlementApproved SettlementStatus = "APPROVED"
	SettlementPaid     SettlementStatus = "PAID"
	SettlementFailed   SettlementStatus = "FAILED"
)

// PayeeDisbursement directs funds to an entity (Claimant, Garage, etc)
type PayeeDisbursement struct {
	PayeeID       string
	Amount        decimal.Decimal
	PaymentMethod string // e.g., TELEBIRR, BANK_TRANSFER
}

// Settlement represents the final calculation and payout configuration
type Settlement struct {
	ID                 string
	ClaimID            string
	Status             SettlementStatus
	GrossLossBase      decimal.Decimal
	PolicyDeductible   decimal.Decimal
	SalvageValueBase   decimal.Decimal
	NetSettlementBase  decimal.Decimal
	Disbursements      []PayeeDisbursement
	CreatedAt          time.Time
}

// NewSettlement computes the advanced deductions logic
func NewSettlement(claimID string, grossLoss, deductible, salvage, policyLimit decimal.Decimal) *Settlement {
	// Formula: MIN(Gross - Deductible - Salvage, PolicyLimit)
	net := grossLoss.Sub(deductible).Sub(salvage)
	
	if net.GreaterThan(policyLimit) {
		net = policyLimit
	}
	if net.LessThan(decimal.Zero) {
		net = decimal.Zero
	}

	return &Settlement{
		ClaimID:           claimID,
		Status:            SettlementPending,
		GrossLossBase:     grossLoss,
		PolicyDeductible:  deductible,
		SalvageValueBase:  salvage,
		NetSettlementBase: net,
		CreatedAt:         time.Now(),
		Disbursements:     make([]PayeeDisbursement, 0),
	}
}

// Approve validates against current reserve and locks for payout
func (s *Settlement) Approve(currentIndemnityReserve decimal.Decimal) error {
	if s.NetSettlementBase.GreaterThan(currentIndemnityReserve) {
		return ErrReserveExceeded
	}
	s.Status = SettlementApproved
	return nil
}
