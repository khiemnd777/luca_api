package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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

		// cache
		field.Int64("order_id").
			Nillable().
			Optional(),

		field.String("order_code").
			Nillable().
			Optional(),

		field.String("process_name").
			Nillable().
			Optional(),

		field.Int("step_number"),

		field.Int64("assigned_id").
			Optional().Nillable(),

		field.String("assigned_name").
			Optional().Nillable(),

		field.String("status").
			Default("pending"), // pending | in_progress | paused | qc | completed | rework | issue

		field.String("color").
			MaxLen(8).
			Optional().
			Nillable(),

		field.String("section_name").
			Optional().
			Nillable(),

		field.JSON("custom_fields", map[string]any{}).
			Optional().
			Default(map[string]any{}),

		field.Time("started_at").
			Optional().Nillable(),

		field.Time("completed_at").
			Optional().Nillable(),

		field.String("note").
			Optional().Nillable(),

		field.Time("updated_at").
			Default(time.Now).UpdateDefault(time.Now),
	}
}

func (OrderItemProcess) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("item", OrderItem.Type).
			Ref("processes").
			Field("order_item_id").
			Required().
			Unique(),
		edge.To("in_progresses", OrderItemProcessInProgress.Type),
	}
}

func (OrderItemProcess) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_item_id", "step_number"),
		index.Fields("order_id", "step_number"),
		index.Fields("assigned_id", "step_number"),
	}
}
