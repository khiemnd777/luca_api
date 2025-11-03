package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// AttributeOption defines the option values for dropdown/checkbox/etc.
type AttributeOption struct {
	ent.Schema
}

func (AttributeOption) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.Int("attribute_id"),
		field.String("option_value").NotEmpty(),
		field.Int("display_order").Default(0),
		field.Bool("deleted").Default(false),
		field.Time("created_at").Default(time.Now),
	}
}

func (AttributeOption) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("attribute", Attribute.Type).
			Ref("options").
			Field("attribute_id").
			Required().
			Unique(),

		edge.From("user", User.Type).
			Ref("attribute_options").
			Field("user_id").
			Required().
			Unique(),

		edge.To("selected_values", AttributeOptionValue.Type),
	}
}
