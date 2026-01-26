package model

import "time"

type AvgTurnaroundResult struct {
	AvgDays   float64
	DeltaDays float64
}

type AvgTurnaroundByDepartment struct {
	DepartmentID int
	AvgDays      float64
}

type CaseDailyStatsUpsert struct {
	ReceivedAt   time.Time
	CompletedAt  time.Time
	DepartmentID int
}
