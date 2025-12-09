package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Category struct {
	ent.Schema
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Optional().
			Nillable(),

		field.Int("collection_id").
			Optional().
			Nillable(),

		field.Bool("active").
			Default(true),

		field.JSON("custom_fields", map[string]any{}).
			Optional().
			Default(map[string]any{}),

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

func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("products", CategoryProduct.Type),
	}
}

func (Category) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "deleted_at"),
		index.Fields("name", "deleted_at"),
		index.Fields("deleted_at"),
	}
}
