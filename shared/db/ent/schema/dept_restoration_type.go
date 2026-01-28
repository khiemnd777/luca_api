package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type RestorationType struct {
	ent.Schema
}

func (RestorationType) Fields() []ent.Field {
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

func (RestorationType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("products", ProductRestorationType.Type),
	}
}

func (RestorationType) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "deleted_at"),
		index.Fields("category_id", "deleted_at"),
		index.Fields("name", "deleted_at"),
		index.Fields("deleted_at"),
	}
}
