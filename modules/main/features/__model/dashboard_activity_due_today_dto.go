package model

import "time"

type DueTodayItem struct {
	ID         int64     `json:"id,omitempty"`
	Code       string    `json:"code,omitempty"`
	Dentist    string    `json:"dentist,omitempty"`
	Patient    string    `json:"patient,omitempty"`
	DeliveryAt time.Time `json:"delivery_at,omitempty"`
	AgeDays    int       `json:"age_days,omitempty"`
	DueType    string    `json:"due_type,omitempty"`
	Priority   string    `json:"priority,omitempty"`
}
