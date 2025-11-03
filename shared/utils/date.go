package utils

import (
	"errors"
	"time"
)

func ParseDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339, // 2025-04-25T15:30:00Z
		"2006-01-02", // 2025-04-25
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.000",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("invalid date format")
}

func ParseNillableDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := ParseDate(s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
