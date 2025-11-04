package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type DepartmentMember struct {
	ent.Schema
}

func (DepartmentMember) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.Int("department_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (DepartmentMember) Edges() []ent.Edge {
	return []ent.Edge{
		// Join -> User
		edge.From("user", User.Type).
			Ref("dept_memberships").
			Field("user_id").
			Unique().
			Required(),

		// Join -> Department
		edge.From("department", Department.Type).
			Ref("members").
			Field("department_id").
			Unique().
			Required(),
	}
}

func (DepartmentMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "department_id").Unique(),
		index.Fields("department_id"),
		index.Fields("user_id"),
		index.Fields("user_id", "created_at"),
	}
}
