package model

import "time"

type ProductDTO struct {
	ID             int            `json:"id,omitempty"`
	Code           *string        `json:"code,omitempty"`
	Name           *string        `json:"name,omitempty"`
	CustomFields   map[string]any `json:"custom_fields,omitempty"`
	ProcessIDs     []int          `json:"process_ids,omitempty"`
	ProcessNames   *string        `json:"process_names,omitempty"`
	BrandNameIDs   []int          `json:"brand_name_ids,omitempty"`
	BrandNameNames *string        `json:"brand_name_names,omitempty"`
	TechniqueIDs   []int          `json:"technique_ids,omitempty"`
	TechniqueNames *string        `json:"technique_names,omitempty"`
	RawMaterialIDs []int          `json:"raw_material_ids,omitempty"`
	RawMaterialNames *string        `json:"raw_material_names,omitempty"`
	RestorationTypeIDs   []int          `json:"restoration_type_ids,omitempty"`
	RestorationTypeNames *string        `json:"restoration_type_names,omitempty"`
	CategoryID     *int           `json:"category_id,omitempty"`
	CategoryName   *string        `json:"category_name,omitempty"`
	// template
	CollectionID *int `json:"collection_id,omitempty"`
	TemplateID   *int `json:"template_id,omitempty"`
	IsTemplate   bool `json:"is_template"`
	// time
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductUpsertDTO struct {
	DTO         ProductDTO `json:"dto"`
	Collections *[]string  `json:"collections,omitempty"`
}
