package model

import (
	"time"
)

type OrderDTO struct {
	// General
	ID           int64          `json:"id,omitempty"`
	Code         *string        `json:"code,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
	CreatedAt    time.Time      `json:"created_at,omitempty"`
	UpdatedAt    time.Time      `json:"updated_at,omitempty"`
	DeliveryDate *time.Time     `json:"delivery_date,omitempty"`
	// Customer
	CustomerID   int64   `json:"customer_id,omitempty"`
	CustomerName *string `json:"customer_name,omitempty"`
	// Latest props
	LatestOrderItemUpsert *OrderItemUpsertDTO `json:"latest_order_item_upsert,omitempty"`
	LatestOrderItem       *OrderItemDTO       `json:"latest_order_item,omitempty"`
	CodeLatest            *string             `json:"code_latest,omitempty"`
	StatusLatest          *string             `json:"status_latest,omitempty"`
	PriorityLatest        *string             `json:"priority_latest,omitempty"`
	// Product props
	ProductID   int      `json:"product_id,omitempty"`
	ProductName *string  `json:"product_name,omitempty"`
	Quantity    *int     `json:"quantity,omitempty"`
	TotalPrice  *float64 `json:"total_price,omitempty"`
	// Remake
	RemakeType  *string `json:"remake_type,omitempty"`
	RemakeCount *int    `json:"remake_count,omitempty"`
}

type OrderUpsertDTO struct {
	DTO         OrderDTO  `json:"dto"`
	Collections *[]string `json:"collections,omitempty"`
}
