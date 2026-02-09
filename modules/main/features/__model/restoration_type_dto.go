package model

import "time"

type RestorationTypeDTO struct {
	ID           int       `json:"id,omitempty"`
	CategoryID   *int      `json:"category_id,omitempty"`
	CategoryName *string   `json:"category_name,omitempty"`
	Name         *string   `json:"name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
