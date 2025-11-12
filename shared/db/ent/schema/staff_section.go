package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type StaffSection struct {
	ent.Schema
}

func (StaffSection) Fields() []ent.Field {
	return []ent.Field{
		field.Int("staff_id"),
		field.Int("section_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (StaffSection) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("staff", Staff.Type).
			Ref("sections").
			Field("staff_id").
			Unique().
			Required(),

		edge.From("section", Section.Type).
			Ref("staffs").
			Field("section_id").
			Unique().
			Required(),
	}
}

func (StaffSection) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("staff_id", "section_id").Unique(),
		index.Fields("staff_id"),
		index.Fields("section_id"),
	}
}
