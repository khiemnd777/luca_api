package utils

import (
	"strings"
	"time"
)

type NullableTime struct {
	*time.Time
}

func (nt *NullableTime) UnmarshalJSON(b []byte) error {
	// Trim quote
	s := strings.Trim(string(b), `"`)

	// empty string → null
	if s == "" || s == "null" {
		nt.Time = nil
		return nil
	}

	// Try parse RFC3339
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try common formats
		layouts := []string{
			"2006-01-02",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z07:00",
		}
		for _, layout := range layouts {
			if tt, err2 := time.Parse(layout, s); err2 == nil {
				nt.Time = &tt
				return nil
			}
		}
		return err
	}

	nt.Time = &t
	return nil
}
