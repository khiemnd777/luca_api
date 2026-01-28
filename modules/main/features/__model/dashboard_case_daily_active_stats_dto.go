package model

import "time"

type ActiveCasesResult struct {
	Value int `json:"value,omitempty"`
	Delta int `json:"delta,omitempty"`
}

type CaseDailyActiveStatsUpsert struct {
	StatAt       time.Time
	DepartmentID int
}
