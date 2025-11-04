package model

import "time"

type DepartmentDTO struct {
	ID          int       `json:"id,omitempty"`
	Slug        string    `json:"slug,omitempty"`
	Active      bool      `json:"active,omitempty"`
	Name        string    `json:"name,omitempty"`
	Logo        *string   `json:"logo,omitempty"`
	Address     *string   `json:"address,omitempty"`
	PhoneNumber *string   `json:"phone_number,omitempty"`
	ParentID    *int      `json:"parent_id,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
