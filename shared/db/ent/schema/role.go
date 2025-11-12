package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Role struct{ ent.Schema }

func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("role_name").
			NotEmpty().
			Unique().
			Comment("Tên vai trò"),
		field.String("display_name").
			Nillable().
			Optional().
			Comment("Tên hiển thị của vai trò"),
		field.String("brief").
			Nillable().
			Optional().
			Comment("Mô tả ngắn gọn về vai trò"),
	}
}

func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("permissions", Permission.Type),
		edge.From("users", User.Type).Ref("roles"),
	}
}
