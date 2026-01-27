package model

import "time"

type ActiveCasesResult struct {
	Value int
	Delta int
}

type CaseDailyActiveStatsUpsert struct {
	StatAt       time.Time
	DepartmentID int
}
