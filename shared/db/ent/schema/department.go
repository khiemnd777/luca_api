package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Department struct {
	ent.Schema
}

func (Department) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Default("Luca"),

		field.String("slug").
			NotEmpty().
			Default("luca"),

		field.String("logo").
			Optional().
			Nillable().
			Comment("Logo (URL)"),

		field.String("address").
			Optional().
			Nillable().
			Comment("Địa chỉ phòng lab"),

		field.String("phone_number").
			Optional().
			Nillable().
			Comment("Số điện thoại"),

		field.Int("parent_id").
			Optional().
			Nillable(),

		field.Bool("active").Default(true),

		field.Bool("deleted").Default(false),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Department) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", Department.Type).
			From("parent").
			Field("parent_id").
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.SetNull, // khi xoá cha, con giữ lại và parent_id = NULL
			}),

		// O2M thực: Department -> DepartmentMember
		edge.To("members", DepartmentMember.Type),
	}
}

func (Department) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug", "deleted"),
		index.Fields("id", "deleted"),
		index.Fields("deleted"),
		index.Fields("parent_id"),
		index.Fields("deleted", "parent_id"),
	}
}
