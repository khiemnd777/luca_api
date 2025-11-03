package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Photo holds the schema definition for the Photo entity.
type Photo struct {
	ent.Schema
}

// Fields of the Photo.
func (Photo) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.Int("folder_id").Optional().Nillable(),
		field.String("url"),
		field.String("provider").Default("default"),
		field.String("name"),
		field.Bool("deleted").Default(false),

		field.String("meta_device").Optional(),
		field.String("meta_os").Optional(),
		field.Float("meta_lat").Optional(),
		field.Float("meta_lng").Optional(),
		field.Int("meta_width").Optional(),
		field.Int("meta_height").Optional(),
		field.Time("meta_captured_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Photo) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("photos").
			Field("user_id").
			Unique().
			Required(),

		edge.From("folder", Folder.Type).
			Ref("photos").
			Field("folder_id").
			Unique(),
	}
}
