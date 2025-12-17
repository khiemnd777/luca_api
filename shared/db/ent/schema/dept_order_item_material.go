package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OrderItemMaterial struct {
	ent.Schema
}

func (OrderItemMaterial) Fields() []ent.Field {
	return []ent.Field{
		field.String("material_code").
			Optional().
			Nillable(),
		field.Int("material_id"),
		field.Int64("order_item_id"),
		field.Int64("order_id"),
		field.Int("quantity").
			Default(1),

		// type: consumable, loaner
		field.String("type").
			MaxLen(16).
			Optional().
			Nillable(),

		// type: returned, on_loan
		field.String("status").
			MaxLen(16).
			Optional().
			Nillable(),

		// type: used for consumable
		field.Float("retail_price").
			Optional().
			Nillable(),
	}
}

func (OrderItemMaterial) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("order_item", OrderItem.Type).
			Required().
			StorageKey(edge.Column("order_item_id")),

		edge.To("order", Order.Type).
			Required().
			StorageKey(edge.Column("order_id")),

		edge.To("material", Material.Type).
			Required().
			StorageKey(edge.Column("material_id")),
	}
}

func (OrderItemMaterial) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_item_id", "material_id").
			Unique(),
		index.Fields("order_id"),
		index.Fields("material_id"),
	}
}
