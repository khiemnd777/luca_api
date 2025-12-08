package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// deprecated
type ProductProcess struct {
	ent.Schema
}

func (ProductProcess) Fields() []ent.Field {
	return []ent.Field{
		field.Int("product_id"),
		field.Int("process_id"),
		// field.String("product_name").
		// 	Optional().
		// 	Nillable(),
		// field.String("section_name").
		// 	Optional().
		// 	Nillable(),
		// field.String("process_name").
		// 	Optional().
		// 	Nillable(),
		// field.String("color").
		// 	MaxLen(8).
		// 	Optional().
		// 	Nillable(),
		// field.Int("display_order").
		// 	Max(100).
		// 	Optional().
		// 	Nillable(),
		field.Time("created_at").Default(time.Now),
	}
}

func (ProductProcess) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("processes").
			Field("product_id").
			Unique().
			Required(),

		edge.From("process", Process.Type).
			Ref("products").
			Field("process_id").
			Unique().
			Required(),
	}
}

func (ProductProcess) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id", "process_id").Unique(),
		index.Fields("process_id"),
		index.Fields("product_id"),
	}
}
