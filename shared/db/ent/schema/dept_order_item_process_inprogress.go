package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OrderItemProcessInProgress struct {
	ent.Schema
}

func (OrderItemProcessInProgress) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			Unique().
			SchemaType(map[string]string{
				"postgres": "bigserial",
			}),
		field.Int64("process_id").Optional().Nillable(),
		field.Int64("prev_process_id").Optional().Nillable(),
		field.Int64("next_process_id").Optional().Nillable(),

		field.String("note").
			Optional().Nillable(),

		// cache
		field.Int64("order_item_id"),
		field.Int64("order_id").
			Nillable().
			Optional(),

		// assignee
		field.Int64("assigned_id").
			Optional().Nillable(),
		field.String("assigned_name").
			Optional().Nillable(),

		// timing
		field.Time("created_at").
			Default(time.Now),
		field.Time("started_at").
			Optional().Nillable(),
		field.Time("completed_at").
			Optional().Nillable(),
		field.Time("updated_at").
			Default(time.Now).UpdateDefault(time.Now),
	}
}

func (OrderItemProcessInProgress) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("process", OrderItemProcess.Type).
			Ref("in_progresses").
			Field("process_id").
			Unique(),
	}
}

func (OrderItemProcessInProgress) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_item_id", "created_at"),
		index.Fields("process_id", "completed_at"),
	}
}
