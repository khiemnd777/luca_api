package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type OrderItemFile struct {
	ent.Schema
}

func (OrderItemFile) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			Unique().
			SchemaType(map[string]string{
				"postgres": "bigserial",
			}),
		field.Int64("order_item_id"),

		field.String("file_url"),
		field.String("file_type").Optional(), // scan_stl | photo | cad | video
		field.String("description").Optional(),

		field.Time("created_at").
			Default(time.Now),
	}
}

func (OrderItemFile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("item", OrderItem.Type).
			Ref("files").
			Field("order_item_id").
			Required().
			Unique(),
	}
}
