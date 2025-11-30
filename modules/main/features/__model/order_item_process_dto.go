package model

import "time"

type OrderItemProcessDTO struct {
	ID           int64      `json:"id,omitempty"`
	OrderID      *int64     `json:"order_id,omitempty"`
	OrderItemID  int64      `json:"order_item_id,omitempty"`
	OrderCode    *string    `json:"order_code,omitempty"`
	ProcessName  *string    `json:"process_name,omitempty"`
	StepNumber   int        `json:"step_number,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Note         *string    `json:"note,omitempty"`
	AssignedID   *int64     `json:"assigned_id,omitempty"`
	AssignedName *string    `json:"assigned_name,omitempty"`
	// CustomFields
	// Status       string         `json:"status,omitempty"`
	CustomFields map[string]any `json:"custom_fields,omitempty"`
}

type OrderItemProcessUpsertDTO struct {
	DTO         OrderItemProcessDTO `json:"dto"`
	Collections *[]string           `json:"collections,omitempty"`
}
