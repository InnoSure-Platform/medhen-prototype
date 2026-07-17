package aggregate

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// UPRSchedule represents the Unearned Premium Reserve amortization schedule.
// For IFRS 17 compliance, premium must be earned proportionally over the coverage period.
type UPRSchedule struct {
	ID                uuid.UUID
	InvoiceID         uuid.UUID
	TotalPremium      decimal.Decimal
	EarnedPremium     decimal.Decimal
	UnearnedPremium   decimal.Decimal
	CoverageStartDate time.Time
	CoverageEndDate   time.Time
	TotalDays         int
	DailyEarnedRate   decimal.Decimal
	LastAmortizedDate time.Time
}

func NewUPRSchedule(invoiceID uuid.UUID, totalPremium decimal.Decimal, start, end time.Time) (*UPRSchedule, error) {
	if !end.After(start) {
		return nil, errors.New("coverage end date must be after start date")
	}

	days := int(end.Sub(start).Hours() / 24)
	if days == 0 {
		days = 1 // Minimum 1 day coverage
	}

	dailyRate := totalPremium.Div(decimal.NewFromInt(int64(days)))

	return &UPRSchedule{
		ID:                uuid.New(),
		InvoiceID:         invoiceID,
		TotalPremium:      totalPremium,
		EarnedPremium:     decimal.Zero,
		UnearnedPremium:   totalPremium,
		CoverageStartDate: start,
		CoverageEndDate:   end,
		TotalDays:         days,
		DailyEarnedRate:   dailyRate,
		LastAmortizedDate: start,
	}, nil
}

// Amortize calculated the earned premium up to the given target date.
func (u *UPRSchedule) Amortize(targetDate time.Time) (earnedThisPeriod decimal.Decimal, err error) {
	if targetDate.Before(u.LastAmortizedDate) {
		return decimal.Zero, errors.New("target date cannot be before last amortized date")
	}

	if targetDate.After(u.CoverageEndDate) {
		targetDate = u.CoverageEndDate
	}

	daysToAmortize := int(targetDate.Sub(u.LastAmortizedDate).Hours() / 24)
	if daysToAmortize <= 0 {
		return decimal.Zero, nil
	}

	earnedAmount := u.DailyEarnedRate.Mul(decimal.NewFromInt(int64(daysToAmortize)))
	
	// Avoid rounding overshoot on the final day
	if u.UnearnedPremium.Sub(earnedAmount).LessThan(decimal.Zero) {
		earnedAmount = u.UnearnedPremium
	}

	u.EarnedPremium = u.EarnedPremium.Add(earnedAmount)
	u.UnearnedPremium = u.UnearnedPremium.Sub(earnedAmount)
	u.LastAmortizedDate = targetDate

	return earnedAmount, nil
}
