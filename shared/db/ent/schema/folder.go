package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Folder holds the schema definition for the Folder entity.
type Folder struct {
	ent.Schema
}

// Fields of the Folder.
func (Folder) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id").Comment("Owner of the folder"),
		field.String("folder_name").NotEmpty().Comment("Name of the folder"),
		field.String("color").Default("#ffffff").Comment("Folder color"),
		field.Bool("shared").Default(false).Comment("Is this folder shared?"),
		field.Int("parent_id").Optional().Nillable().Comment("Parent folder ID, if any"),
		field.Bool("deleted").Default(false).Comment("Soft delete flag"),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Folder.
func (Folder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("folders").
			Field("user_id").
			Unique().
			Required(),

		edge.To("children", Folder.Type).
			From("parent").
			Unique().
			Field("parent_id"),

		edge.To("photos", Photo.Type),
	}
}
