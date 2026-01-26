package model

import "time"

type AvgRemakeResult struct {
	Rate      float64 // 0.031 = 3.1%
	DeltaRate float64
}

type CaseDailyRemakeStatsUpsert struct {
	CompletedAt  time.Time
	DepartmentID int
	IsRemake     bool
}
