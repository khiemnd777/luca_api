package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// deprecated
type CategoryProduct struct {
	ent.Schema
}

func (CategoryProduct) Fields() []ent.Field {
	return []ent.Field{
		field.Int("category_id"),
		field.Int("product_id"),
		field.Int("display_order").
			Max(100).
			Optional().
			Nillable(),
		field.Time("created_at").Default(time.Now),
	}
}

func (CategoryProduct) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("category", Category.Type).
			Ref("products").
			Field("category_id").
			Unique().
			Required(),

		edge.From("product", Product.Type).
			Ref("categories").
			Field("product_id").
			Unique().
			Required(),
	}
}

func (CategoryProduct) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id", "category_id").Unique(),
	}
}
