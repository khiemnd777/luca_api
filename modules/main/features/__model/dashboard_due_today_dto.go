package model

import "time"

type DueTodayItem struct {
	Code       string    `json:"code,omitempty"`
	Dentist    string    `json:"dentist,omitempty"`
	Patient    string    `json:"patient,omitempty"`
	DeliveryAt time.Time `json:"delivery_at,omitempty"`
	Priority   string    `json:"priority,omitempty"`
}
