package model

import "time"

type OrderDTO struct {
	// General
	ID           int64          `json:"id,omitempty"`
	Code         *string        `json:"code,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	CreatedAt    time.Time      `json:"created_at,omitempty"`
	UpdatedAt    time.Time      `json:"updated_at,omitempty"`
	// Customer
	CustomerID   int64   `json:"customer_id,omitempty"`
	CustomerName *string `json:"customer_name,omitempty"`
	// Latest Props
	LatestOrderItemUpsert *OrderItemUpsertDTO `json:"latest_order_item_upsert,omitempty"`
	LatestOrderItem       *OrderItemDTO       `json:"latest_order_item,omitempty"`
	CodeLatest            *string             `json:"code_latest,omitempty"`
	StatusLatest          *string             `json:"status_latest,omitempty"`
	PriorityLatest        *string             `json:"priority_latest,omitempty"`
}

type OrderUpsertDTO struct {
	DTO         OrderDTO  `json:"dto"`
	Collections *[]string `json:"collections,omitempty"`
}
