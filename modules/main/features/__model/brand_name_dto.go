package model

import "time"

type BrandNameDTO struct {
	ID         int       `json:"id,omitempty"`
	CategoryID *int      `json:"category_id,omitempty"`
	Name       *string   `json:"name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
