package model

import "time"

type OrderItemDTO struct {
	ID           int64          `json:"id,omitempty"`
	OrderID      int64          `json:"order_id,omitempty"`
	ParentItemID *int64         `json:"parent_item_id,omitempty"`
	Code         *string        `json:"code,omitempty"`
	CodeOriginal *string        `json:"code_original,omitempty"`
	RemakeCount  int            `json:"remake_count,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	CreatedAt    time.Time      `json:"created_at,omitempty"`
	UpdatedAt    time.Time      `json:"updated_at,omitempty"`
}

type OrderItemUpsertDTO struct {
	DTO         OrderItemDTO `json:"dto"`
	Collections *[]string    `json:"collections,omitempty"`
}
