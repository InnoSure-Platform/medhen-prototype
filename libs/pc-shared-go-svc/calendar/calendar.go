package calendar

import (
	"time"
)

// GeezDate represents a date in the Ethiopian (Ge'ez) calendar.
type GeezDate struct {
	Year  int
	Month int
	Day   int
}

// Converter defines the interface for converting between Gregorian and Ge'ez calendars.
type Converter interface {
	ToGeez(gregorian time.Time) GeezDate
	ToGregorian(geez GeezDate) time.Time
}

// JDNConverter provides an exact Julian Day Number implementation.
type JDNConverter struct{}

func NewConverter() Converter {
	return &JDNConverter{}
}

// ToGeez converts a Gregorian time.Time to a Ge'ez date using precise Julian Day mathematics.
func (c *JDNConverter) ToGeez(gregorian time.Time) GeezDate {
	jdn := gregorianToJDN(gregorian.Year(), int(gregorian.Month()), gregorian.Day())
	year, month, day := jdnToGeez(jdn)
	return GeezDate{Year: year, Month: month, Day: day}
}

// ToGregorian converts a Ge'ez date to a Gregorian time.Time using precise Julian Day mathematics.
func (c *JDNConverter) ToGregorian(geez GeezDate) time.Time {
	jdn := geezToJDN(geez.Year, geez.Month, geez.Day)
	year, month, day := jdnToGregorian(jdn)
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// --- Exact Mathematical Conversions ---

// gregorianToJDN converts Gregorian (Year, Month, Day) to Julian Day Number.
func gregorianToJDN(year, month, day int) int {
	a := (14 - month) / 12
	y := year + 4800 - a
	m := month + 12*a - 3
	jdn := day + (153*m+2)/5 + 365*y + y/4 - y/100 + y/400 - 32045
	return jdn
}

// jdnToGregorian converts a Julian Day Number to Gregorian (Year, Month, Day).
func jdnToGregorian(jdn int) (year, month, day int) {
	a := jdn + 32044
	b := (4*a + 3) / 146097
	c := a - (146097*b)/4
	d := (4*c + 3) / 1461
	e := c - (1461*d)/4
	m := (5*e + 2) / 153
	day = e - (153*m+2)/5 + 1
	month = m + 3 - 12*(m/10)
	year = b*100 + d - 4800 + (m / 10)
	return
}

// geezToJDN converts Ethiopian (Ge'ez) to Julian Day Number.
// The Ethiopian calendar epoch (Year 1, Month 1, Day 1) is JDN 1723856.
func geezToJDN(year, month, day int) int {
	return (year-1)*365 + year/4 + 30*(month-1) + day + 1723855
}

// jdnToGeez converts Julian Day Number to Ethiopian (Ge'ez).
func jdnToGeez(jdn int) (year, month, day int) {
	r := (jdn - 1723856) % 1461
	n := r%365 + 365*(r/1460)
	year = 4*((jdn-1723856)/1461) + r/365 - r/1460 + 1
	month = n/30 + 1
	day = n%30 + 1
	return
}
