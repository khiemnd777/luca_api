package model

import (
	"time"
)

type PatientDTO struct {
	ID          int          `json:"id,omitempty"`
	Name        string       `json:"name,omitempty"`
	PhoneNumber *string      `json:"phone_number,omitempty"`
	Brief       *string      `json:"brief,omitempty"`
	Active      bool         `json:"active,omitempty"`
	Clinics     []*ClinicDTO `json:"clinics,omitempty"`
	ClinicIDs   []int        `json:"clinic_ids,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
