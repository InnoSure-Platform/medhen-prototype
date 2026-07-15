// Package calendar provides Ethiopian (Ge'ez) calendar conversion helpers for display.
// Storage remains Gregorian UTC; presentation layers convert for bilingual docs/UI.
package calendar

import (
	"fmt"
	"time"
)

// EthiopianDate is a date in the Ethiopian calendar.
type EthiopianDate struct {
	Year  int
	Month int
	Day   int
}

var ethMonthsEN = []string{
	"", "Meskerem", "Tikimt", "Hidar", "Tahsas", "Tir", "Yekatit",
	"Megabit", "Miazia", "Ginbot", "Sene", "Hamle", "Nehase", "Pagumen",
}

var ethMonthsAM = []string{
	"", "መስከረም", "ጥቅምት", "ኅዳር", "ታኅሣሥ", "ጥር", "የካቲት",
	"መጋቢት", "ሚያዝያ", "ግንቦት", "ሰኔ", "ሐምሌ", "ነሐሴ", "ጳጉሜን",
}

// FromGregorian converts a Gregorian date to Ethiopian (algorithm per ethio-calendar norms).
func FromGregorian(t time.Time) EthiopianDate {
	// Use date components in local civil time of the instant (UTC date for storage/display consistency in Phase 0).
	y, m, d := t.UTC().Date()
	return gregorianToEthiopian(y, int(m), d)
}

func gregorianToEthiopian(gy, gm, gd int) EthiopianDate {
	// Based on the standard JD offset algorithm used across Ethiopian calendar libraries.
	jOffset := 1723856
	a := (14 - gm) / 12
	y := gy + 4800 - a
	m := gm + 12*a - 3
	jdn := gd + (153*m+2)/5 + 365*y + y/4 - y/100 + y/400 - 32045
	r := jdn - jOffset
	ey := r / 365
	rem := r % 365
	em := rem/30 + 1
	ed := rem%30 + 1
	return EthiopianDate{Year: ey, Month: em, Day: ed}
}

func (e EthiopianDate) FormatEN() string {
	name := "?"
	if e.Month >= 1 && e.Month <= 13 {
		name = ethMonthsEN[e.Month]
	}
	return fmt.Sprintf("%d %s %d", e.Day, name, e.Year)
}

func (e EthiopianDate) FormatAM() string {
	name := "?"
	if e.Month >= 1 && e.Month <= 13 {
		name = ethMonthsAM[e.Month]
	}
	return fmt.Sprintf("%d %s %d", e.Day, name, e.Year)
}
