package model

import "time"

type OrderItemProcessInProgressAndProcessDTO struct {
	ID           int64      `json:"id,omitempty"`
	Note         *string    `json:"note,omitempty"`
	AssignedID   *int64     `json:"assigned_id,omitempty"`
	AssignedName *string    `json:"assigned_name,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ProcessName  *string    `json:"process_name,omitempty"`
	SectionName  *string    `json:"section_name,omitempty"`
	Color        *string    `json:"color,omitempty"`
}
