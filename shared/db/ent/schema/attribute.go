package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Attribute holds the schema definition for the Attribute entity.
type Attribute struct {
	ent.Schema
}

func (Attribute) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.String("attribute_name").NotEmpty(),
		field.String("attribute_type").Default("text"), // "text", "dropdown", "checkbox", "multicheckbox", "linkedlist".
		field.Bool("deleted").Default(false),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Attribute) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("attributes").
			Field("user_id").
			Required().
			Unique(),

		edge.To("options", AttributeOption.Type),
		edge.To("option_values", AttributeOptionValue.Type),
	}
}
