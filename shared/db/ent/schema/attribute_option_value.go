package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// AttributeOptionValue links a product to a selected option of an attribute.
type AttributeOptionValue struct {
	ent.Schema
}

func (AttributeOptionValue) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.Int("attribute_id"),
		field.Int("option_id"),
		field.Bool("deleted").Default(false),
		field.Time("created_at").Default(time.Now),
	}
}

func (AttributeOptionValue) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("attribute_option_values").
			Field("user_id").
			Required().
			Unique(),

		edge.From("attribute", Attribute.Type).
			Ref("option_values").
			Field("attribute_id").
			Required().
			Unique(),

		edge.From("attribute_option", AttributeOption.Type).
			Ref("selected_values").
			Field("option_id").
			Required().
			Unique(),
	}
}
