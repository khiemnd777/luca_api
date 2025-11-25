package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OrderItem struct {
	ent.Schema
}

func (OrderItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			Unique().
			SchemaType(map[string]string{
				"postgres": "bigserial",
			}),

		field.Int64("order_id"),

		field.Int64("parent_item_id").
			Optional().
			Nillable(),

		field.String("code").
			Optional().
			Nillable(),

		field.String("code_original").
			Optional().
			Nillable(),

		field.String("remake_type").
			Optional().Nillable(), // adjust | remake

		field.String("remake_reason").
			Optional().Nillable(),

		field.Int("remake_count").
			Default(0),

		// product info
		// field.String("product_type"), // crown | veneer | ...
		// field.Int64("material_id").Optional(),
		// field.String("shade").Optional(),

		field.Strings("tooth_positions").Optional(),

		// field.String("category_slug"),

		field.JSON("custom_fields", map[string]any{}).
			Default(map[string]any{}),

		field.String("status").
			Default("pending"), // pending | in_progress | qc | completed | rework | issue

		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).UpdateDefault(time.Now),

		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

func (OrderItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("items").
			Field("order_id").
			Required().
			Unique(),

		edge.To("processes", OrderItemProcess.Type),

		edge.To("files", OrderItemFile.Type),

		edge.To("remake_logs", OrderItemRemakeLog.Type),

		edge.From("parent", OrderItem.Type).
			Ref("children").
			Field("parent_item_id").
			Unique(),

		edge.To("children", OrderItem.Type),
	}
}

func (OrderItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "deleted_at"),
		index.Fields("code"),
		index.Fields("code_original"),
		index.Fields("code", "deleted_at").Unique(),
		index.Fields("code_original", "created_at", "deleted_at"),
		index.Fields("order_id", "created_at", "deleted_at"),
		index.Fields("deleted_at"),
	}
}
