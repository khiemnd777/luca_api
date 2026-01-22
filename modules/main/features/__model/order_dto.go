package model

import (
	"time"
)

type OrderDTO struct {
	// General
	ID              int64          `json:"id,omitempty"`
	Code            *string        `json:"code,omitempty"`
	PromotionCodeID *int           `json:"promotion_code_id,omitempty"`
	PromotionCode   *string        `json:"promotion_code,omitempty"`
	CustomFields    map[string]any `json:"custom_fields,omitempty"`
	CreatedAt       time.Time      `json:"created_at,omitempty"`
	UpdatedAt       time.Time      `json:"updated_at,omitempty"`
	DeliveryDate    *time.Time     `json:"delivery_date,omitempty"`
	// Customer
	// deprecated
	CustomerID *int64 `json:"customer_id,omitempty"`
	// deprecated
	CustomerName *string `json:"customer_name,omitempty"`
	ClinicID     *int    `json:"clinic_id,omitempty"`
	ClinicName   *string `json:"clinic_name,omitempty"`
	DentistID    *int    `json:"dentist_id,omitempty"`
	DentistName  *string `json:"dentist_name,omitempty"`
	PatientID    *int    `json:"patient_id,omitempty"`
	PatientName  *string `json:"patient_name,omitempty"`
	// Latest props
	LatestOrderItemUpsert *OrderItemUpsertDTO `json:"latest_order_item_upsert,omitempty"`
	LatestOrderItem       *OrderItemDTO       `json:"latest_order_item,omitempty"`
	CodeLatest            *string             `json:"code_latest,omitempty"`
	StatusLatest          *string             `json:"status_latest,omitempty"`
	PriorityLatest        *string             `json:"priority_latest,omitempty"`
	ProcessIDLatest       *int                `json:"process_id_latest,omitempty"`
	ProcessNameLatest     *string             `json:"process_name_latest,omitempty"`
	SectionNameLatest     *string             `json:"section_name_latest,omitempty"`
	LeaderIDLatest        *int                `json:"leader_id_latest,omitempty"`
	LeaderNameLatest      *string             `json:"leader_name_latest,omitempty"`

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
