package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Section defines the schema for company/lab divisions or departments.
type Section struct {
	ent.Schema
}

// Fields of the Section.
func (Section) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty(),

		field.String("code").
			Optional().
			Nillable().
			MaxLen(50),

		field.String("description").
			Optional().
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

func (Section) Edges() []ent.Edge {
	return nil
}

func (Section) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
	}
}
