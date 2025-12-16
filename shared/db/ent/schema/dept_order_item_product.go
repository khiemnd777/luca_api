package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OrderItemProduct struct {
	ent.Schema
}

func (OrderItemProduct) Fields() []ent.Field {
	return []ent.Field{
		field.String("product_code").
			Optional().
			Nillable(),
		field.Int("product_id"),
		field.Int64("order_item_id"),
		field.Int64("order_id"),
		field.Int("quantity").
			Default(1),
		field.Float("retail_price").
			Optional().
			Nillable(),
	}
}

func (OrderItemProduct) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("order_item", OrderItem.Type).
			Required().
			StorageKey(edge.Column("order_item_id")),

		edge.To("order", Order.Type).
			Required().
			StorageKey(edge.Column("order_id")),

		edge.To("product", Product.Type).
			Required().
			StorageKey(edge.Column("product_id")),
	}
}

func (OrderItemProduct) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_item_id", "product_id").
			Unique(),
		index.Fields("order_id"),
		index.Fields("product_id"),
	}
}
