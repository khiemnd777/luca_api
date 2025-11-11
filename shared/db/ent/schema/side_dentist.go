package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Dentist struct {
	ent.Schema
}

func (Dentist) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty(),

		field.String("phone_number").
			Optional().
			Nillable(),

		field.String("brief").
			Optional().
			Nillable().
			MaxLen(300),

		field.Bool("active").
			Default(true),

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

func (Dentist) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("clinics", ClinicDentist.Type),
	}
}

func (Dentist) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "deleted_at"),
		index.Fields("deleted_at"),
	}
}
