package calendar

import (
	"fmt"
	"time"
)

// Duration measures the number of calendar months and/or days
// between two calendar dates, depending on how it is constructed.
type Duration struct {
	Days int
	Months int
}

// String renders a Duration in "1m2d" format.
func (d Duration) String() string {
	if d.Days == 0 && d.Months == 0 {
		return "0m0d"
	}
	if d.Days == 0 {
		return fmt.Sprintf("%dm", d.Months)
	}
	if d.Months == 0 {
		return fmt.Sprintf("%dd", d.Days)
	}
	return fmt.Sprintf("%dm%dd", d.Months, d.Days)
}

// DurationDays ignores the clock time and compares the dates a and b.
// It returns a Duration representing the number of days between them.
func DurationDays(a, b time.Time) Duration {
	var d Duration
	if b.Before(a) {
		d = DurationDays(b, a)
		return Duration{Days: -d.Days}
	}
	aYear, bYear := a.Year(), b.Year()
	if aYear < bYear {
		d.Days += epochDays(bYear) - epochDays(aYear)
	}
	d.Days += b.YearDay() - a.YearDay()
	return d
}

// epochDays returns the number of days, including leap days,
// since the fictitious date Jan 1, 0001 until Jan 1, y.
// This uses the proleptic Gregorian calendar.
func epochDays(y int) int {
	// Since Feb 29 hasn't happened in y yet,
	// only sum leap days until y-1.
	prev := y-1
	return prev*365 + (prev/4) - (prev/100) + (prev/400)
}

// DurationMonthDays builds a Duration while attempting
// to align days of the month.
func DurationMonthDays(a, b time.Time) Duration {
	var d Duration
	d.Months = 12 * (b.Year()-a.Year()) + int(b.Month()-a.Month())
	d.Days = b.Day() - a.Day()
	return d
}

// After returns the time.Time which is d after t.
func (d Duration) After(t time.Time) time.Time {
	return t.AddDate(0, d.Months, d.Days)
}

// Before returns the time.Time which is d before t.
func (d Duration) Before(t time.Time) time.Time {
	return t.AddDate(0, -d.Months, -d.Days)
}

// Add adds o to d.
func (d Duration) Add(o Duration) Duration {
	d.Months += o.Months
	d.Days += o.Days
	return d
}

// Sub subtracts o from d.
func (d Duration) Sub(o Duration) Duration {
	d.Months -= o.Months
	d.Days -= o.Days
	return d
}

// Mul multiplies the Duration by s.
func (d Duration) Mul(s int) Duration {
	d.Days *= s
	d.Months *= s
	return d
}
