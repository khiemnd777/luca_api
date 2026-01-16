package schema

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type PromotionCondition struct {
	ent.Schema
}

func (PromotionCondition) Fields() []ent.Field {
	return []ent.Field{
		field.Int("promo_code_id"),
		field.Enum("condition_type").
			Values("ORDER_IS_REMAKE", "REMAKE_COUNT_LTE", "REMAKE_WITHIN_DAYS", "REMAKE_REASON").
			Immutable(),
		field.JSON("condition_value", json.RawMessage{}).
			Optional().
			Immutable(),
	}
}

func (PromotionCondition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("promotion_code", PromotionCode.Type).
			Ref("conditions").
			Field("promo_code_id").
			Required().
			Unique(),
	}
}
