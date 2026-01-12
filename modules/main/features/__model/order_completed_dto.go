package model

import "time"

type CompletedOrderDTO struct {
	ID             int64     `json:"id,omitempty"`
	Code           *string   `json:"code,omitempty"`
	CodeLatest     *string   `json:"code_latest,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	StatusLatest   *string   `json:"status_latest,omitempty"`
	PriorityLatest *string   `json:"priority_latest,omitempty"`
}
