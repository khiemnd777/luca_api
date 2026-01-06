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
		field.Int64("original_order_item_id").
			Optional().
			Nillable(),
		field.Int64("order_id"),
		field.Int("quantity").
			Default(1),

		field.Float("retail_price").
			Optional().
			Nillable(),

		field.Bool("is_cloneable").
			Optional().
			Nillable(),

		field.String("teeth_position").
			Optional().
			Nillable(),

		field.String("note").
			Optional().
			Nillable(),
	}
}

func (OrderItemProduct) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order_item", OrderItem.Type).
			Ref("products").
			Field("order_item_id").
			Unique().
			Required(),

		edge.From("product", Product.Type).
			Ref("order_items").
			Field("product_id").
			Unique().
			Required(),
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
