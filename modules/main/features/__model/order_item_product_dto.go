package model

type OrderItemProductDTO struct {
	ID                  int      `json:"id,omitempty"`
	ProductCode         *string  `json:"product_code,omitempty"`
	ProductName         *string  `json:"product_name,omitempty"`
	ProductID           int      `json:"product_id,omitempty"`
	OrderItemID         int64    `json:"order_item_id,omitempty"`
	OriginalOrderItemID *int64   `json:"original_order_item_id,omitempty"`
	OrderItemCode       *string  `json:"order_item_code,omitempty"`
	OrderID             int64    `json:"order_id,omitempty"`
	Quantity            int      `json:"quantity,omitempty"`
	RetailPrice         *float64 `json:"retail_price,omitempty"`
	IsCloneable         *bool    `json:"is_cloneable,omitempty"`
}
