package model

import "time"

type CompletedCasesResult struct {
	Value int `json:"value,omitempty"`
	Delta int `json:"delta,omitempty"`
}

type CaseDailyCompletedStatsUpsert struct {
	CompletedAt  time.Time
	DepartmentID int
}
