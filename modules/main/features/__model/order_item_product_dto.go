package model

type OrderItemProductDTO struct {
	ID          int      `json:"id,omitempty"`
	ProductCode *string  `json:"product_code,omitempty"`
	ProductID   int      `json:"product_id,omitempty"`
	OrderItemID int64    `json:"order_item_id,omitempty"`
	OrderID     int64    `json:"order_id,omitempty"`
	Quantity    int      `json:"quantity,omitempty"`
	RetailPrice *float64 `json:"retail_price,omitempty"`
}
