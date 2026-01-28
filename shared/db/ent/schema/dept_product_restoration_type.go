package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ProductRestorationType struct {
	ent.Schema
}

func (ProductRestorationType) Fields() []ent.Field {
	return []ent.Field{
		field.Int("product_id"),
		field.Int("restoration_type_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (ProductRestorationType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("restoration_types").
			Field("product_id").
			Unique().
			Required(),

		edge.From("restoration_type", RestorationType.Type).
			Ref("products").
			Field("restoration_type_id").
			Unique().
			Required(),
	}
}

func (ProductRestorationType) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id", "restoration_type_id").Unique(),
		index.Fields("restoration_type_id"),
		index.Fields("product_id"),
	}
}
