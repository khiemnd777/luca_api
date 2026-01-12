package model

import "time"

type InProcessOrderDTO struct {
	ID             int64      `json:"id,omitempty"`
	Code           *string    `json:"code,omitempty"`
	CodeLatest     *string    `json:"code_latest,omitempty"`
	DeliveryDate   *time.Time `json:"delivery_date,omitempty"`
	TotalPrice     *float64   `json:"total_price,omitempty"`
	StatusLatest   *string    `json:"status_latest,omitempty"`
	PriorityLatest *string    `json:"priority_latest,omitempty"`
}
