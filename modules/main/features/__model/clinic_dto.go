package model

import (
	"time"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
)

type ClinicDTO struct {
	ID         int                 `json:"id,omitempty"`
	Name       string              `json:"name,omitempty"`
	Brief      *string             `json:"brief,omitempty"`
	Logo       *string             `json:"logo,omitempty"`
	Active     bool                `json:"active,omitempty"`
	Dentists   []*model.DentistDTO `json:"dentists,omitempty"`
	DentistIDs []int               `json:"dentist_ids,omitempty"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}
