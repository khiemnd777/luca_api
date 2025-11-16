package model

import (
	"time"
)

type ClinicDTO struct {
	ID           int            `json:"id,omitempty"`
	Name         string         `json:"name,omitempty"`
	Address      *string        `json:"address,omitempty"`
	PhoneNumber  *string        `json:"phone_number,omitempty"`
	Brief        *string        `json:"brief,omitempty"`
	Logo         *string        `json:"logo,omitempty"`
	Active       bool           `json:"active,omitempty"`
	Dentists     []*DentistDTO  `json:"dentists,omitempty"`
	DentistIDs   []int          `json:"dentist_ids,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
