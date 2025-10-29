package utils

import (
	"fmt"
	"time"
)

func ParseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q (want MM-YYYY): %w", s, err)
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func FormatMonthYear(t time.Time) string { return t.Format("01-2006") }
