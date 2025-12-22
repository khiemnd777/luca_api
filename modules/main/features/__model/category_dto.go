package model

import "time"

type CategoryDTO struct {
	ID              int            `json:"id,omitempty"`
	Level           int            `json:"level,omitempty"`
	ParentID        *int           `json:"parent_id,omitempty"`
	CategoryIDLv1   *int           `json:"category_id_lv1,omitempty"`
	CategoryNameLv1 *string        `json:"category_name_lv1,omitempty"`
	CategoryIDLv2   *int           `json:"category_id_lv2,omitempty"`
	CategoryNameLv2 *string        `json:"category_name_lv2,omitempty"`
	CategoryIDLv3   *int           `json:"category_id_lv3,omitempty"`
	CategoryNameLv3 *string        `json:"category_name_lv3,omitempty"`
	Name            *string        `json:"name,omitempty"`
	CollectionID    *int           `json:"collection_id,omitempty"`
	Active          bool           `json:"active,omitempty"`
	ProductIDs      []int          `json:"product_ids,omitempty"`
	ProcessIDs      []int          `json:"process_ids,omitempty"`
	CustomFields    map[string]any `json:"custom_fields,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type CategoryUpsertDTO struct {
	DTO         CategoryDTO `json:"dto"`
	Collections *[]string   `json:"collections,omitempty"`
}
