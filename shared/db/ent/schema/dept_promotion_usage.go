package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type PromotionUsage struct {
	ent.Schema
}

func (PromotionUsage) Fields() []ent.Field {
	return []ent.Field{
		field.Int("promo_code_id"),
		field.Int("order_id"),
		field.Int("user_id"),
		field.Int("discount_amount"),
		field.Time("used_at").
			Default(time.Now),
	}
}

func (PromotionUsage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("promotion_code", PromotionCode.Type).
			Ref("usages").
			Field("promo_code_id").
			Required().
			Unique(),
	}
}
