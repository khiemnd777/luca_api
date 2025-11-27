package model

import "time"

type ProductDTO struct {
	ID           int            `json:"id,omitempty"`
	Code         *string        `json:"code,omitempty"`
	Name         *string        `json:"name,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	ProcessIDs   []int          `json:"process_ids,omitempty"`
	ProcessNames *string        `json:"process_names,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type ProductShortDTO struct {
	ID           int            `json:"id,omitempty"`
	Code         *string        `json:"code,omitempty"`
	Name         *string        `json:"name,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
}

type ProductSearchDTO struct {
	ID   int     `json:"id,omitempty"`
	Code *string `json:"code,omitempty"`
	Name *string `json:"name,omitempty"`
}

type ProductUpsertDTO struct {
	DTO         ProductDTO `json:"dto"`
	Collections *[]string  `json:"collections,omitempty"`
}
