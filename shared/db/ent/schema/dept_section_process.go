package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type SectionProcess struct {
	ent.Schema
}

func (SectionProcess) Fields() []ent.Field {
	return []ent.Field{
		field.Int("section_id"),
		field.Int("process_id"),
		field.String("section_name").
			Optional().
			Nillable(),
		field.String("process_name").
			Optional().
			Nillable(),
		field.String("color").
			MaxLen(8).
			Optional().
			Nillable(),
		field.Int("display_order").
			Max(100).
			Optional().
			Nillable(),
		field.Time("created_at").Default(time.Now),
	}
}

func (SectionProcess) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("section", Section.Type).
			Ref("processes").
			Field("section_id").
			Unique().
			Required(),

		edge.From("process", Process.Type).
			Ref("sections").
			Field("process_id").
			Unique().
			Required(),
	}
}

func (SectionProcess) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("section_id", "process_id").Unique(),
	}
}
