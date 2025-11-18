package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Supplier struct {
	ent.Schema
}

func (Supplier) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			Optional().
			Nillable().
			Unique(),

		field.String("name").
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

func (Supplier) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("materials", MaterialSupplier.Type),
	}
}

func (Supplier) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "deleted_at"),
		index.Fields("code"),
		index.Fields("code", "deleted_at"),
		index.Fields("name", "deleted_at"),
		index.Fields("deleted_at"),
	}
}
