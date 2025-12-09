package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Order struct {
	ent.Schema
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			Unique().
			SchemaType(map[string]string{
				"postgres": "bigserial",
			}),

		field.String("code").
			Optional().
			Nillable(),

		field.JSON("custom_fields", map[string]any{}).
			Optional().
			Default(map[string]any{}),

		field.Int64("customer_id").Optional(),
		field.String("customer_name").
			Nillable().
			Optional(),

		// Cache & Table
		field.String("code_latest").
			Optional().
			Nillable(),

		field.String("status_latest").
			Optional().
			Nillable(),

		field.String("priority_latest").
			Optional().
			Nillable(),

		field.Int("product_id").Optional(),
		field.String("product_name").
			Nillable().
			Optional(),

		field.Int("quantity").
			Nillable().
			Optional(),

		field.Float("total_price").
			Nillable().
			Optional(),

		field.Time("delivery_date").
			Nillable().
			Optional(),

		field.String("remake_type").
			Nillable().
			Optional(),

		field.Int("remake_count").
			Nillable().
			Optional(),

		// times
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("items", OrderItem.Type),
	}
}

func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "deleted_at"),
		index.Fields("code"),
		index.Fields("code", "deleted_at").Unique(),
		index.Fields("deleted_at"),
	}
}
