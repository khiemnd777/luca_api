package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type PromotionCode struct {
	ent.Schema
}

func (PromotionCode) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			Unique(),
		field.String("name").
			Optional().
			Nillable(),
		field.Enum("discount_type").
			Values("fixed", "percent", "free_shipping"),
		field.Int("discount_value"),
		field.Int("max_discount_amount").
			Optional().
			Nillable(),
		field.Int("min_order_value").
			Optional().
			Nillable(),
		field.Int("total_usage_limit").
			Optional().
			Nillable(),
		field.Int("usage_per_user").
			Optional().
			Nillable(),
		field.Time("start_at"),
		field.Time("end_at"),
		field.Bool("is_active").
			Default(true),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (PromotionCode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("scopes", PromotionScope.Type),
		edge.To("conditions", PromotionCondition.Type),
		edge.To("usages", PromotionUsage.Type),
	}
}
