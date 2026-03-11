package model

import "errors"

var ErrInvalidOrExpiredOrderCode = errors.New("invalid or expired order_code")

var (
	ErrInvalidDeliveryQRToken      = errors.New("invalid delivery qr token")
	ErrDeliveryQRTokenAlreadyUsed  = errors.New("delivery qr token already used")
	ErrOrderAlreadyDelivered       = errors.New("order already delivered")
	ErrDeliveryQRSessionExpired    = errors.New("delivery qr session expired")
	ErrDeliveryQRSessionNotFound   = errors.New("delivery qr session not found")
	ErrDeliveryQRConfirmConcurrent = errors.New("delivery qr confirm failed due to concurrent update")
)
