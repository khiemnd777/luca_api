package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type OrderItemProcess struct {
	ent.Schema
}

func (OrderItemProcess) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			Unique().
			SchemaType(map[string]string{
				"postgres": "bigserial",
			}),
		field.Int64("order_item_id"),

		field.String("process_name"),
		field.Int("step_number"),

		field.Int64("assigned_to").
			Optional().Nillable(),

		field.String("status").
			Default("pending"), // pending | in_progress | paused | qc | completed | rework | issue

		field.Time("started_at").
			Optional().Nillable(),

		field.Time("completed_at").
			Optional().Nillable(),

		field.String("note").
			Optional().Nillable(),
	}
}

func (OrderItemProcess) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("item", OrderItem.Type).
			Ref("processes").
			Field("order_item_id").
			Required().
			Unique(),
	}
}
