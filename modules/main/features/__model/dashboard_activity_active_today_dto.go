package model

import "time"

type ActiveTodayItem struct {
	ID         int64     `json:"id,omitempty"`
	Code       string    `json:"code,omitempty"`
	Dentist    string    `json:"dentist,omitempty"`
	Patient    string    `json:"patient,omitempty"`
	DeliveryAt time.Time `json:"delivery_at,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	AgeDays    int       `json:"age_days,omitempty"`
	Priority   *string   `json:"priority,omitempty"`
}
