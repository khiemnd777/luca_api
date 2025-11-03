package cart_types

type OrderItemPayload struct {
	ProductID       int     `json:"product_id"`
	PublisherID     int     `json:"publisher_id"`
	Price           float64 `json:"price"`
	Quantity        int     `json:"quantity"`
	PreOrderPercent float64 `json:"pre_order_percent"`
	Tags            *string `json:"tags"`
}
