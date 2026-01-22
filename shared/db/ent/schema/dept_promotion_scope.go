package schema

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type PromotionScope struct {
	ent.Schema
}

func (PromotionScope) Fields() []ent.Field {
	return []ent.Field{
		field.Int("promo_code_id"),
		field.Enum("scope_type").
			Values("ALL", "USER", "SELLER", "CATEGORY", "PRODUCT").
			Immutable(),
		field.JSON("scope_value", json.RawMessage{}).
			Optional().
			Immutable(),
	}
}

func (PromotionScope) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("promotion_code", PromotionCode.Type).
			Ref("scopes").
			Field("promo_code_id").
			Required().
			Unique(),
	}
}
