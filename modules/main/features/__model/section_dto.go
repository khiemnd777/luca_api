package model

import "time"

type SectionDTO struct {
	ID           int            `json:"id,omitempty"`
	DepartmentID int            `json:"department_id,omitempty"`
	Name         string         `json:"name,omitempty"`
	Code         *string        `json:"code,omitempty"`
	Description  string         `json:"description,omitempty"`
	Active       bool           `json:"active,omitempty"`
	Color        *string        `json:"color,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	CreatedAt    time.Time      `json:"created_at,omitempty"`
	UpdatedAt    time.Time      `json:"updated_at,omitempty"`
	DeletedAt    *time.Time     `json:"deleted_at,omitempty"`
	// Processes
	ProcessIDs   []int   `json:"process_ids,omitempty"`
	ProcessNames *string `json:"process_names,omitempty"`
}
