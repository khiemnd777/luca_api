package model

import "time"

type OrderItemProcessInProgressDTO struct {
	ID              int64      `json:"id,omitempty"`
	ProcessID       *int64     `json:"process_id,omitempty"`
	PrevProcessID   *int64     `json:"prev_process_id,omitempty"`
	NextProcessID   *int64     `json:"next_process_id,omitempty"`
	NextProcessName *string    `json:"next_process_name,omitempty"`
	NextSectionName *string    `json:"next_section_name,omitempty"`
	NextLeaderID    *int       `json:"next_leader_id,omitempty"`
	NextLeaderName  *string    `json:"next_leader_name,omitempty"`
	CheckInNote     *string    `json:"check_in_note,omitempty"`
	CheckOutNote    *string    `json:"check_out_note,omitempty"`
	OrderItemID     int64      `json:"order_item_id,omitempty"`
	OrderID         *int64     `json:"order_id,omitempty"`
	OrderItemCode   *string    `json:"order_item_code,omitempty"`
	SectionName     *string    `json:"section_name,omitempty"`
	AssignedID      *int64     `json:"assigned_id,omitempty"`
	AssignedName    *string    `json:"assigned_name,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at,omitempty"`
}
