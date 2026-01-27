package model

import "time"

type CompletedCasesResult struct {
	Value int
	Delta int
}

type CaseDailyCompletedStatsUpsert struct {
	CompletedAt  time.Time
	DepartmentID int
}
