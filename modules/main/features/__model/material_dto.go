package model

import "time"

type MaterialDTO struct {
	ID            int            `json:"id,omitempty"`
	Code          *string        `json:"code,omitempty"`
	Name          *string        `json:"name,omitempty"`
	SupplierIDs   []int          `json:"supplier_ids,omitempty"`
	SupplierNames []string       `json:"supplier_names,omitempty"`
	CustomFields  map[string]any `json:"custom_fields,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
