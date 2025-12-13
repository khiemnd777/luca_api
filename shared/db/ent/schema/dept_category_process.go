package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// deprecated
type CategoryProcess struct {
	ent.Schema
}

func (CategoryProcess) Fields() []ent.Field {
	return []ent.Field{
		field.Int("category_id"),
		field.Int("process_id"),
		field.Int("display_order").
			Max(100).
			Optional().
			Nillable(),
		field.Time("created_at").Default(time.Now),
	}
}

func (CategoryProcess) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("category", Category.Type).
			Ref("processes").
			Field("category_id").
			Unique().
			Required(),

		edge.From("process", Process.Type).
			Ref("categories").
			Field("process_id").
			Unique().
			Required(),
	}
}

func (CategoryProcess) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("category_id", "process_id").Unique(),
	}
}
