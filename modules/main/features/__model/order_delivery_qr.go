package model

import "time"

type DeliveryQRSession struct {
	SessionID     string    `json:"session_id"`
	OrderID       int       `json:"order_id"`
	OrderCode     string    `json:"order_code"`
	OrderItemCode string    `json:"order_item_code"`
	QRTokenID     int       `json:"qr_token_id"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type OrderDeliveryAuditAction string

const (
	OrderDeliveryAuditActionScan    OrderDeliveryAuditAction = "scan"
	OrderDeliveryAuditActionConfirm OrderDeliveryAuditAction = "confirm"
	OrderDeliveryAuditActionExpired OrderDeliveryAuditAction = "expired"
	OrderDeliveryAuditActionInvalid OrderDeliveryAuditAction = "invalid"
)

type OrderDeliveryQRMessageType string

const (
	OrderDeliveryQRMessageTypeSessionStarted OrderDeliveryQRMessageType = "DeliverySessionStarted"
)
