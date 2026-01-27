package model

import "time"

type AvgRemakeResult struct {
	Rate      float64 `json:"rate,omitempty"` // 0.031 = 3.1%
	DeltaRate float64 `json:"delta_rate,omitempty"`
}

type CaseDailyRemakeStatsUpsert struct {
	CompletedAt  time.Time
	DepartmentID int
	IsRemake     bool
}
