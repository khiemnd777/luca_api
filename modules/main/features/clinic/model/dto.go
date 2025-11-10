package model

import "time"

type ClinicDTO struct {
	ID        int        `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Brief     *string    `json:"brief,omitempty"`
	Logo      *string    `json:"logo,omitempty"`
	Active    bool       `json:"active,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
