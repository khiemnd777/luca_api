package model

import "time"

type OrderItemProductDTO struct {
	ID          int      `json:"id,omitempty"`
	ProductID   int      `json:"product_id,omitempty"`
	OrderItemID int64    `json:"order_item_id,omitempty"`
	OrderID     int64    `json:"order_id,omitempty"`
	Quantity    int      `json:"quantity,omitempty"`
	RetailPrice *float64 `json:"retail_price,omitempty"`
}

type OrderItemDTO struct {
	// general
	ID           int64          `json:"id,omitempty"`
	OrderID      int64          `json:"order_id,omitempty"`
	ParentItemID *int64         `json:"parent_item_id,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	CreatedAt    time.Time      `json:"created_at,omitempty"`
	UpdatedAt    time.Time      `json:"updated_at,omitempty"`
	// order
	Code         *string  `json:"code,omitempty"`
	CodeOriginal *string  `json:"code_original,omitempty"`
	RemakeCount  int      `json:"remake_count,omitempty"`
	TotalPrice   *float64 `json:"total_price,omitempty"`
	// product
	Products []*OrderItemProductDTO `json:"products,omitempty"`
	// processes
	OrderItemProcesses []*OrderItemProcessDTO `json:"order_item_processes,omitempty"`
}

type OrderItemUpsertDTO struct {
	DTO         OrderItemDTO `json:"dto"`
	Collections *[]string    `json:"collections,omitempty"`
}

type OrderItemHistoricalDTO struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	IsCurrent   bool      `json:"is_current"`
	IsHighlight bool      `json:"is_highlight"`
}
