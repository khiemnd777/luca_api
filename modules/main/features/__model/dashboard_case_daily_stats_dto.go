package model

import "time"

type AvgTurnaroundResult struct {
	AvgDays   float64 `json:"avg_days,omitempty"`
	DeltaDays float64 `json:"delta_days,omitempty"`
}

type CaseDailyStatsUpsert struct {
	ReceivedAt   time.Time
	CompletedAt  time.Time
	DepartmentID int
}
