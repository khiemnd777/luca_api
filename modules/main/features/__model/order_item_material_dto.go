package model

type OrderItemMaterialDTO struct {
	ID            int      `json:"id,omitempty"`
	MaterialCode  *string  `json:"material_code,omitempty"`
	MaterialName  *string  `json:"material_name,omitempty"`
	MaterialID    int      `json:"material_id,omitempty"`
	OrderItemID   int64    `json:"order_item_id,omitempty"`
	OrderItemCode *string  `json:"order_item_code,omitempty"`
	OrderID       int64    `json:"order_id,omitempty"`
	Quantity      int      `json:"quantity,omitempty"`
	Type          *string  `json:"type,omitempty"`
	Status        *string  `json:"status,omitempty"`
	RetailPrice   *float64 `json:"retail_price,omitempty"`
	IsCloneable   *bool    `json:"is_cloneable,omitempty"`
}
