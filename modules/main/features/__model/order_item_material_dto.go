package model

import "time"

type OrderItemMaterialDTO struct {
	ID                  int        `json:"id,omitempty"`
	MaterialCode        *string    `json:"material_code,omitempty"`
	MaterialName        *string    `json:"material_name,omitempty"`
	MaterialID          int        `json:"material_id,omitempty"`
	OrderItemID         int64      `json:"order_item_id,omitempty"`
	OriginalOrderItemID *int64     `json:"original_order_item_id,omitempty"`
	OrderItemCode       *string    `json:"order_item_code,omitempty"`
	OrderID             int64      `json:"order_id,omitempty"`
	Quantity            int        `json:"quantity,omitempty"`
	Type                *string    `json:"type,omitempty"`
	Status              *string    `json:"status,omitempty"`
	RetailPrice         *float64   `json:"retail_price,omitempty"`
	IsCloneable         *bool      `json:"is_cloneable,omitempty"`
	ClinicID            *int       `json:"clinic_id,omitempty"`
	ClinicName          *string    `json:"clinic_name,omitempty"`
	DentistID           *int       `json:"dentist_id,omitempty"`
	DentistName         *string    `json:"dentist_name,omitempty"`
	PatientID           *int       `json:"patient_id,omitempty"`
	PatientName         *string    `json:"patient_name,omitempty"`
	OnLoanAt            *time.Time `json:"on_loan_at,omitempty"`
	ReturnedAt          *time.Time `json:"returned_at,omitempty"`
	Note                *string    `json:"note,omitempty"`
}
