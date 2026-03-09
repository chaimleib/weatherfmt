package calendar_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chaimleib/weatherfmt/calendar"
)

func TestDuration_String(t *testing.T) {
	cases := []struct{
		D calendar.Duration
		Want string
	}{
		{
			D: calendar.Duration{Months: 0, Days: 9},
			Want: "9d",
		},
		{
			D: calendar.Duration{Months: 2, Days: 9},
			Want: "2m9d",
		},
		{
			D: calendar.Duration{Months: -1, Days: 9},
			Want: "-1m9d",
		},
		{
			D: calendar.Duration{Months: 1, Days: -9},
			Want: "1m-9d",
		},
		{
			D: calendar.Duration{Months: 1, Days: 0},
			Want: "1m",
		},
		{
			D: calendar.Duration{},
			Want: "0m0d",
		},
		{
			D: calendar.Duration{Months: 1, Days: -21},
			Want: "1m-21d",
		},
		{
			D: calendar.Duration{Months: 25, Days: -30},
			Want: "25m-30d",
		},
		{
			D: calendar.Duration{Months: 12, Days: 0},
			Want: "12m",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%+v.String() = %s", c.D, c.Want)
		t.Run(name, func(t *testing.T) {
			got := c.D.String()
			if got != c.Want {
				t.Errorf("want %q, got %q", c.Want, got)
			}
		})
	}
}
func date(s string) time.Time {
	formats := []string{
		time.DateOnly,
		time.RFC3339,
	}
	var errs []error
	for _, f := range formats {
		t, err := time.Parse(f, s)
		if err != nil {
			errs = append(errs, err)
		} else {
			return t
		}
	}
	panic(fmt.Errorf(
		"failed to parse %q: %v",
		s,
		errors.Join(errs...),
	))
}

func TestDurationDays(t *testing.T) {
	cases := []struct{
		A, B string
		Want int
	}{
		{A: "2026-01-01", B: "2026-01-10", Want: 9},
		{
			A: "2026-01-01",
			B: "2026-03-10",
			Want: 68,
		},
		{A: "2026-02-01", B: "2026-01-10", Want: -22},
		{A: "2026-01-10", B: "2026-02-01", Want: 22},
		{A: "2026-01-10", B: "2026-02-10", Want: 31},
		{A: "2025-10-08T11:00:00Z", B: "2025-10-08T22:00:00Z", Want: 0},
		{A: "2025-12-31", B: "2026-01-10", Want: 10},
		{A: "1999-12-31", B: "2002-01-01", Want: 732},
		{A: "2003-01-01", B: "2004-01-01", Want: 365},
		{A: "2004-01-01", B: "2005-01-01", Want: 366},
	}
	for _, c := range cases {
		name := fmt.Sprintf("DurationDays(%q, %q) = %v", c.A, c.B, c.Want)
		t.Run(name, func(t *testing.T) {
			a, b := date(c.A), date(c.B)
			got := calendar.DurationDays(a, b)
			if got.Months != 0 {
				t.Errorf("want zero months, got %+v", got)
			}
			if got.Days != c.Want {
				t.Errorf("want %d days, got %+v", c.Want, got)
			}
		})
	}
}

func TestDurationMonthDays(t *testing.T) {
	cases := []struct{
		A, B string
		Want calendar.Duration
	}{
		{
			A: "2026-01-01",
			B: "2026-01-10",
			Want: calendar.Duration{Months: 0, Days: 9},
		},
		{
			A: "2026-01-01",
			B: "2026-03-10",
			Want: calendar.Duration{Months: 2, Days: 9},
		},
		{
			A: "2026-02-01",
			B: "2026-01-10",
			Want: calendar.Duration{Months: -1, Days: 9},
		},
		{
			A: "2026-01-10",
			B: "2026-02-01",
			Want: calendar.Duration{Months: 1, Days: -9},
		},
		{
			A: "2026-01-10",
			B: "2026-02-10",
			Want: calendar.Duration{Months: 1, Days: 0},
		},
		{
			A: "2025-10-08T11:00:00Z",
			B: "2025-10-08T22:00:00Z",
			Want: calendar.Duration{},
		},
		{
			A: "2025-12-31",
			B: "2026-01-10",
			Want: calendar.Duration{Months: 1, Days: -21},
		},
		{
			A: "1999-12-31",
			B: "2002-01-01",
			Want: calendar.Duration{Months: 25, Days: -30},
		},
		{
			A: "2003-01-01",
			B: "2004-01-01",
			Want: calendar.Duration{Months: 12, Days: 0},
		},
		{
			A: "2004-01-01",
			B: "2005-01-01",
			Want: calendar.Duration{Months: 12, Days: 0},
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("DurationMonthsDays(%q, %q) = %+v", c.A, c.B, c.Want)
		t.Run(name, func(t *testing.T) {
			a, b := date(c.A), date(c.B)
			got := calendar.DurationMonthDays(a, b)
			if got.Months != c.Want.Months {
				t.Errorf("want %d months, got %+v", c.Want.Months, got)
			}
			if got.Days != c.Want.Days {
				t.Errorf("want %d days, got %+v", c.Want.Days, got)
			}
		})
	}
}

func TestDuration_After_Before(t *testing.T) {
	cases := []struct{
		D calendar.Duration
		A string
		B string
	}{
		{
			D: calendar.Duration{Days: 9},
			A: "2026-01-01",
			B: "2026-01-10",
		},
		{
			D: calendar.Duration{Months: 2, Days: 9},
			A: "2026-01-01",
			B: "2026-03-10",
		},
		{
			D: calendar.Duration{Months: -1, Days: 9},
			A: "2026-02-01",
			B: "2026-01-10",
		},
		{
			D: calendar.Duration{Months: 1, Days: -9},
			A: "2026-01-10",
			B: "2026-02-01",
		},
		{
			D: calendar.Duration{Months: 1, Days: 0},
			A: "2026-01-10",
			B: "2026-02-10",
		},
		{
			D: calendar.Duration{},
			A: "2025-10-08",
			B: "2025-10-08",
		},
		{
			D: calendar.Duration{Months: 1, Days: -21},
			A: "2025-12-31",
			B: "2026-01-10",
		},
		{
			D: calendar.Duration{Months: 25, Days: -30},
			A: "1999-12-31",
			B: "2002-01-01",
		},
		{
			D: calendar.Duration{Months: 12, Days: 0},
			A: "2003-01-01",
			B: "2004-01-01",
		},
		{
			D: calendar.Duration{Months: 12, Days: 0},
			A: "2004-01-01",
			B: "2005-01-01",
		},
	}
	for _, c := range cases {
		name := fmt.Sprintf("%+v.After(%q) = %q", c.D, c.A, c.B)
		t.Run(name, func(t *testing.T) {
			tm, want := date(c.A), date(c.B)
			got := c.D.After(tm)
			if !got.Equal(want) {
				t.Errorf("want %s, got %s", c.B, got)
			}
		})

		name = fmt.Sprintf("%+v.Before(%q) = %q", c.D, c.B, c.A)
		t.Run(name, func(t *testing.T) {
			tm, want := date(c.B), date(c.A)
			got := c.D.Before(tm)
			if !got.Equal(want) {
				t.Errorf("want %s, got %s", c.B, got)
			}
		})
	}
}

func parseDuration(s string) calendar.Duration {
	var m, d int
	var err error
	var ok bool
	rest := s

	if rest, ok = strings.CutSuffix(rest, "d"); ok {
		dStr := lastNum(rest)
		if d, err = strconv.Atoi(dStr); err != nil {
			panic(fmt.Errorf("error parsing %q at %q: %w", s, rest, err))
		}
		rest = rest[:len(rest)-len(dStr)]
	}
	
	if rest, ok = strings.CutSuffix(rest, "m"); ok {
		mStr := lastNum(rest)
		if m, err = strconv.Atoi(mStr); err != nil {
			panic(fmt.Errorf("error parsing %q at %q: %w", s, rest, err))
		}
		rest = rest[:len(rest)-len(mStr)]
	}

	if rest != "" {
		panic(fmt.Errorf(
			"error parsing %q: leftover duration string: %q",
			s,
			rest,
		))
	}
	return calendar.Duration{Months: m, Days: d}
}

func lastNum(s string) string {
	var i int
	for i = len(s)-1; i >= 0; i-- {
		if i < len(s)-1 && s[i] == '-' {
			return s[i:]
		}
		if s[i] < '0' || s[i] > '9' {
			return s[i+1:]
		}
	}
	return s
}

func TestDuration_Add_Sub(t *testing.T) {
	cases := []struct{
		X, Y, Z string
	}{
		{"1d", "1d", "2d"},
		{"1d", "2d", "3d"},
		{"1d", "-2d", "-1d"},
		{"2m", "3m", "5m"},
		{"1m4d", "2m11d", "3m15d"},
		{"1m", "2d", "1m2d"},
		{"1m", "-1m", "0m0d"},
	}
	for _, c := range cases {
		x, y, z := parseDuration(c.X), parseDuration(c.Y), parseDuration(c.Z)
		name := fmt.Sprintf("%s + %s = %s", c.X, c.Y, c.Z)
		t.Run(name, func(t *testing.T) {
			got := x.Add(y)
			if got != z {
				t.Errorf("want %+v, got %+v", z, got)
			}
		})

		name = fmt.Sprintf("%s - %s = %s", c.Z, c.X, c.Y)
		t.Run(name, func(t *testing.T) {
			got := z.Sub(x)
			if got != y {
				t.Errorf("want %+v, got %+v", y, got)
			}
		})
	}
}

func TestDuration_Mul(t *testing.T) {
	cases := []struct{
		K int
		D, Want string
	}{
		{1, "1d", "1d"},
		{3, "1d", "3d"},
		{-2, "1d", "-2d"},
		{4, "2m", "8m"},
		{3, "1m4d", "3m12d"},
	}
	for _, c := range cases {
		d, want := parseDuration(c.D), parseDuration(c.Want)
		name := fmt.Sprintf("%d*%s = %s", c.K, c.D, c.Want)
		t.Run(name, func(t *testing.T) {
			got := d.Mul(c.K)
			if got != want {
				t.Errorf("want %+v, got %+v", want, got)
			}
		})
	}
}
