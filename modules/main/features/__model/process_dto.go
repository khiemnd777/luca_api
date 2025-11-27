package model

import "time"

type ProcessDTO struct {
	ID           int            `json:"id,omitempty"`
	Code         *string        `json:"code,omitempty"`
	Name         *string        `json:"name,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
