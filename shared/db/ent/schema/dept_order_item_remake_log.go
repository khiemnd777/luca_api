package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type OrderItemRemakeLog struct {
	ent.Schema
}

func (OrderItemRemakeLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			Unique().
			SchemaType(map[string]string{
				"postgres": "bigserial",
			}),
		field.Int64("item_id"),
		field.String("action"), // adjust | remake
		field.String("reason").Optional(),

		field.Int64("by_user").
			Optional().Nillable(),

		field.Time("created_at").
			Default(time.Now),
	}
}

func (OrderItemRemakeLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("item", OrderItem.Type).
			Ref("remake_logs").
			Field("item_id").
			Required().
			Unique(),
	}
}
