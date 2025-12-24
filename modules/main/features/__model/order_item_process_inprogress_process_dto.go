package model

import "time"

type OrderItemProcessInProgressAndProcessDTO struct {
	ID            int64      `json:"id,omitempty"`
	OrderID       *int64     `json:"order_id,omitempty"`
	OrderItemID   int64      `json:"order_item_id,omitempty"`
	OrderItemCode *string    `json:"order_item_code,omitempty"`
	CheckInNote   *string    `json:"check_in_note,omitempty"`
	CheckOutNote  *string    `json:"check_out_note,omitempty"`
	AssignedID    *int64     `json:"assigned_id,omitempty"`
	AssignedName  *string    `json:"assigned_name,omitempty"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	ProcessName   *string    `json:"process_name,omitempty"`
	SectionName   *string    `json:"section_name,omitempty"`
	Color         *string    `json:"color,omitempty"`
}
