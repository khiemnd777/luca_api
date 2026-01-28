package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type RawMaterial struct {
	ent.Schema
}

func (RawMaterial) Fields() []ent.Field {
	return []ent.Field{
		field.Int("category_id").
			Optional().
			Nillable(),

		field.String("name").
			Optional().
			Nillable(),

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

func (RawMaterial) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("products", ProductRawMaterial.Type),
	}
}

func (RawMaterial) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "deleted_at"),
		index.Fields("category_id", "deleted_at"),
		index.Fields("name", "deleted_at"),
		index.Fields("deleted_at"),
	}
}
