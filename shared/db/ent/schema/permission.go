package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Permission struct{ ent.Schema }

func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.String("permission_name").
			NotEmpty().
			Comment("Tên phân quyền, ví dụ: 'Product Read'"),
		field.String("permission_value").
			NotEmpty().
			Unique().
			Comment("Giá trị phân quyền, ví dụ: 'product_read'"),
	}
}

func (Permission) Edges() []ent.Edge {
	return []ent.Edge{
		// Inverse of Role.permissions
		edge.From("roles", Role.Type).Ref("permissions"),
	}
}
