package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Staff struct {
	ent.Schema
}

func (Staff) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Staff) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("staff").
			Unique().
			Required(),

		edge.To("sections", StaffSection.Type),
	}
}
