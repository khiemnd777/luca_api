package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OrderDeliveryProof struct {
	ent.Schema
}

func (OrderDeliveryProof) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Immutable().
			Unique().
			SchemaType(map[string]string{
				"postgres": "serial",
			}),
		field.Int64("order_id"),
		field.Int64("order_item_id").
			Optional().
			Nillable(),
		field.Int("qr_token_id"),
		field.String("image_url"),
		field.Int64("image_size"),
		field.String("image_mime_type"),
		field.Time("created_at").
			Default(time.Now),
	}
}

func (OrderDeliveryProof) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("delivery_proofs").
			Field("order_id").
			Required().
			Unique(),
		edge.From("qr_token", OrderDeliveryQRToken.Type).
			Ref("delivery_proof").
			Field("qr_token_id").
			Required().
			Unique(),
	}
}

func (OrderDeliveryProof) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("order_item_id"),
		index.Fields("qr_token_id").Unique(),
	}
}
