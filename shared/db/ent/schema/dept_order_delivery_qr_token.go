package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OrderDeliveryQRToken struct {
	ent.Schema
}

func (OrderDeliveryQRToken) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Immutable().
			Unique().
			SchemaType(map[string]string{
				"postgres": "serial",
			}),
		field.Int64("order_id"),
		field.String("token_hash"),
		field.Bool("used").Default(false),
		field.Time("used_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (OrderDeliveryQRToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("delivery_qr_tokens").
			Field("order_id").
			Required().
			Unique(),
		edge.To("audit_logs", OrderDeliveryAuditLog.Type),
		edge.To("delivery_proof", OrderDeliveryProof.Type).
			Unique(),
	}
}

func (OrderDeliveryQRToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token_hash").Unique(),
		index.Fields("order_id", "used"),
	}
}
