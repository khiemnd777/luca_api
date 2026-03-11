package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OrderDeliveryAuditLog struct {
	ent.Schema
}

func (OrderDeliveryAuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Immutable().
			Unique().
			SchemaType(map[string]string{
				"postgres": "serial",
			}),
		field.Int64("order_id"),
		field.Int("qr_token_id").
			Optional().
			Nillable(),
		field.String("action"),
		field.String("ip").
			Optional().
			Nillable(),
		field.String("user_agent").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (OrderDeliveryAuditLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("delivery_audit_logs").
			Field("order_id").
			Required().
			Unique(),
		edge.From("qr_token", OrderDeliveryQRToken.Type).
			Ref("audit_logs").
			Field("qr_token_id").
			Unique(),
	}
}

func (OrderDeliveryAuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id", "created_at"),
		index.Fields("qr_token_id"),
		index.Fields("action"),
	}
}
