package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Product struct {
	ent.Schema
}

func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			Optional().
			Nillable(),

		field.String("name").
			Optional().
			Nillable(),

		// base product
		field.Bool("is_default").
			Default(true),

		field.Int("template_id").
			Optional().
			Nillable(),

		field.Bool("is_template").Default(false),

		field.Int("collection_id").
			Optional().
			Nillable(),

		// activated
		field.Bool("active").
			Default(true),

		// custom fields
		field.JSON("custom_fields", map[string]any{}).
			Optional().
			Default(map[string]any{}),

		// category
		field.Int("category_id").
			Optional().
			Nillable(),

		field.String("category_name").
			Optional().
			Nillable(),

		// cache
		field.String("process_names").
			Optional().
			Nillable(),

		// times
		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("processes", ProductProcess.Type),
		edge.To("categories", CategoryProduct.Type),
		edge.To("order_item_products", OrderItemProduct.Type),
	}
}

func (Product) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "deleted_at"),
		index.Fields("code"),
		index.Fields("code", "deleted_at").Unique(),
		index.Fields("name", "deleted_at"),
		index.Fields("deleted_at"),
	}
}
