package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
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
		field.Int("department_id"),

		field.String("name").
			NotEmpty(),

		field.String("code").
			Optional().
			Nillable().
			MaxLen(50),

		field.String("description").
			Optional().
			MaxLen(300),

		field.String("color").
			MaxLen(8).
			Optional().
			Nillable(),

		field.JSON("custom_fields", map[string]any{}).
			Optional().
			Default(map[string]any{}),

		// cache
		field.String("process_names").
			Optional().
			Nillable(),

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
	return []ent.Edge{
		edge.From("department", Department.Type).
			Ref("sections").
			Field("department_id").
			Unique().
			Required(),

		edge.To("processes", SectionProcess.Type),

		edge.To("staffs", StaffSection.Type),
	}
}

func (Section) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
	}
}
