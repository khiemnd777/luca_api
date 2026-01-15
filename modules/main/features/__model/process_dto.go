package model

import "time"

type ProcessDTO struct {
	ID           int            `json:"id,omitempty"`
	Code         *string        `json:"code,omitempty"`
	Name         *string        `json:"name,omitempty"`
	Color        *string        `json:"color,omitempty"`
	SectionID    *int           `json:"section_id,omitempty"`
	SectionName  *string        `json:"section_name,omitempty"`
	LeaderID     *int           `json:"leader_id,omitempty"`
	LeaderName   *string        `json:"leader_name,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
