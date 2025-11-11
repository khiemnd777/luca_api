package model

import "time"

type SectionDTO struct {
	ID          int        `json:"id,omitempty"`
	Name        string     `json:"name,omitempty"`
	Code        *string    `json:"code,omitempty"`
	Description string     `json:"description,omitempty"`
	Active      bool       `json:"active,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
