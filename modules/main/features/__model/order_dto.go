package model

import "time"

type OrderDTO struct {
	ID              int64               `json:"id,omitempty"`
	Code            *string             `json:"code,omitempty"`
	CustomerID      int64               `json:"customer_id,omitempty"`
	CustomerName    *string             `json:"customer_name,omitempty"`
	Priority        int                 `json:"priority,omitempty"`
	Status          string              `json:"status,omitempty"`
	CustomFields    map[string]any      `json:"custom_fields,omitempty"`
	LatestOrderItem *OrderItemUpsertDTO `json:"latest_order_item_dto,omitempty"`
	CreatedAt       time.Time           `json:"created_at,omitempty"`
	UpdatedAt       time.Time           `json:"updated_at,omitempty"`
}

type OrderUpsertDTO struct {
	DTO         OrderDTO  `json:"dto"`
	Collections *[]string `json:"collections,omitempty"`
}
