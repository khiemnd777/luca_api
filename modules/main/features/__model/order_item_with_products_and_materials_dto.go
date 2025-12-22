package model

type OrderItemWithProductsAndMaterialsDTO struct {
	ID                  int64                   `json:"id,omitempty"`
	OrderID             int64                   `json:"order_id,omitempty"`
	Products            []*OrderItemProductDTO  `json:"products,omitempty"`
	ConsumableMaterials []*OrderItemMaterialDTO `json:"consumable_materials,omitempty"`
	LoanerMaterials     []*OrderItemMaterialDTO `json:"loaner_materials,omitempty"`
}

type OrderProductsAndMaterialsDTO []OrderItemWithProductsAndMaterialsDTO
