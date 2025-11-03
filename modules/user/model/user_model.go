package model

type BankQRCodeModel struct {
	ID         int    `json:"id"`
	BankQRCode string `json:"bank_qr_code"`
	Name       string `json:"name"`
}

type QRCodeModel struct {
	ID     int     `json:"id"`
	QRCode string  `json:"qr_code"`
	Name   string  `json:"name"`
	Avatar *string `json:"avatar"`
}
