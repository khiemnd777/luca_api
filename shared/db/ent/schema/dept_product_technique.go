package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ProductTechnique struct {
	ent.Schema
}

func (ProductTechnique) Fields() []ent.Field {
	return []ent.Field{
		field.Int("product_id"),
		field.Int("technique_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (ProductTechnique) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("techniques").
			Field("product_id").
			Unique().
			Required(),

		edge.From("technique", Technique.Type).
			Ref("products").
			Field("technique_id").
			Unique().
			Required(),
	}
}

func (ProductTechnique) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id", "technique_id").Unique(),
		index.Fields("technique_id"),
		index.Fields("product_id"),
	}
}
