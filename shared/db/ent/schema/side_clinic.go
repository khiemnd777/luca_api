package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Clinic struct {
	ent.Schema
}

func (Clinic) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty(),

		field.String("address").
			Optional().
			Nillable(),

		field.String("phone_number").
			Optional().
			Nillable(),

		field.String("brief").
			Optional().
			Nillable().
			MaxLen(300),

		field.String("logo").
			Optional().
			Nillable(),

		field.Bool("active").
			Default(true),

		// JSONB cho custom fields
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

func (Clinic) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("dentists", ClinicDentist.Type),
	}
}

func (Clinic) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "deleted_at"),
		index.Fields("deleted_at"),
	}
}
