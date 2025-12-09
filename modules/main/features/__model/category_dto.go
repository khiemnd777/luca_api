package model

import "time"

type CategoryDTO struct {
	ID           int            `json:"id,omitempty"`
	Name         *string        `json:"name,omitempty"`
	CollectionID *int           `json:"collection_id,omitempty"`
	Active       bool           `json:"active,omitempty"`
	ProductIDs   []int          `json:"product_ids,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type CategoryUpsertDTO struct {
	DTO         CategoryDTO `json:"dto"`
	Collections *[]string   `json:"collections,omitempty"`
}
