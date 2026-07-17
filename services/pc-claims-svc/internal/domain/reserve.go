package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrNegativeBalance   = errors.New("reserve balance cannot be negative")
	ErrAuthorityExceeded = errors.New("authority limit exceeded")
)

// MultiCurrencyAmount is the foundational value object for all financials
type MultiCurrencyAmount struct {
	Amount       decimal.Decimal
	Currency     string
	BaseAmount   decimal.Decimal
	ExchangeRate decimal.Decimal
}

// ReserveType defines the ledger partition
type ReserveType string

const (
	Indemnity ReserveType = "INDEMNITY"
	Expense   ReserveType = "EXPENSE"
	Recovery  ReserveType = "RECOVERY"
)

// ReserveEntry is an immutable ledger line item
type ReserveEntry struct {
	ID                 int64
	ClaimID            string
	ReserveType        ReserveType
	TransactionType    string // SET, INCREASE, DECREASE, PAYMENT_DRAWDOWN
	Amount             MultiCurrencyAmount
	RunningBalanceBase decimal.Decimal
	ReasonCode         string
	AuthorID           string
	
	// Bi-Temporal Versioning
	SystemValidFrom    time.Time
	SystemValidTo      time.Time
	BusinessValidFrom  time.Time
	BusinessValidTo    time.Time
}

// SubrogationNetting tracks fractional recoveries
type SubrogationNetting struct {
	TotalRecoveryBase decimal.Decimal
	EICRetentionBase  decimal.Decimal
	ReinsurerShareBase decimal.Decimal
}

// ReserveLedger is the aggregate root for tracking financials of a claim
type ReserveLedger struct {
	ClaimID   string
	Entries   []ReserveEntry
	Balances  map[ReserveType]decimal.Decimal // current base amounts
}

func NewReserveLedger(claimID string) *ReserveLedger {
	return &ReserveLedger{
		ClaimID:  claimID,
		Entries:  make([]ReserveEntry, 0),
		Balances: make(map[ReserveType]decimal.Decimal),
	}
}

// AdjustReserve appends a new entry ensuring the balance never drops below zero
func (l *ReserveLedger) AdjustReserve(entry ReserveEntry, authorityLimit decimal.Decimal) error {
	if entry.Amount.BaseAmount.GreaterThan(authorityLimit) {
		return ErrAuthorityExceeded
	}

	currentBalance := l.Balances[entry.ReserveType]
	
	// Determine impact based on transaction type
	var newBalance decimal.Decimal
	if entry.TransactionType == "INCREASE" || entry.TransactionType == "SET" {
		newBalance = currentBalance.Add(entry.Amount.BaseAmount)
	} else if entry.TransactionType == "DECREASE" || entry.TransactionType == "PAYMENT_DRAWDOWN" {
		newBalance = currentBalance.Sub(entry.Amount.BaseAmount)
	}

	if newBalance.LessThan(decimal.Zero) {
		return ErrNegativeBalance
	}

	entry.RunningBalanceBase = newBalance
	
	// Set default Bi-Temporal bounds if missing
	now := time.Now().UTC()
	if entry.SystemValidFrom.IsZero() {
		entry.SystemValidFrom = now
	}
	entry.SystemValidTo = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	if entry.BusinessValidFrom.IsZero() {
		entry.BusinessValidFrom = now
	}
	entry.BusinessValidTo = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

	l.Entries = append(l.Entries, entry)
	l.Balances[entry.ReserveType] = newBalance

	return nil
}

// CalculateFractionalRecovery splits a subrogation receipt
func (l *ReserveLedger) CalculateFractionalRecovery(grossRecoveryBase decimal.Decimal, reinsurerPercentage decimal.Decimal) SubrogationNetting {
	reinsurerShare := grossRecoveryBase.Mul(reinsurerPercentage).Round(2)
	eicRetention := grossRecoveryBase.Sub(reinsurerShare)

	return SubrogationNetting{
		TotalRecoveryBase:  grossRecoveryBase,
		EICRetentionBase:   eicRetention,
		ReinsurerShareBase: reinsurerShare,
	}
}
