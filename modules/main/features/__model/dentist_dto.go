package model

import (
	"time"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
)

type DentistDTO struct {
	ID        int                `json:"id,omitempty"`
	Name      string             `json:"name,omitempty"`
	Brief     *string            `json:"brief,omitempty"`
	Active    bool               `json:"active,omitempty"`
	Clinics   []*model.ClinicDTO `json:"clinics,omitempty"`
	ClinicIDs []int              `json:"clinic_ids,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}
