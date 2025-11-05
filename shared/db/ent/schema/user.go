package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").Optional().Unique(),
		field.String("password").Sensitive(),
		field.String("name").Default(""),
		field.String("phone").Optional().Unique(),
		field.Bool("active").Default(true),
		field.String("avatar").Optional(),   // Avatar URL from Google, Facebook, ...
		field.String("provider").Optional(), // 'google', 'facebook', ...
		field.String("provider_id").Optional(),
		field.String("ref_code").Optional().Nillable(), // Ref. code
		field.String("qr_code").Optional().Nillable(),  // User QR code
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("roles", Role.Type),
		edge.To("refresh_tokens", RefreshToken.Type),
		edge.To("folders", Folder.Type),
		edge.To("photos", Photo.Type),
		edge.To("attributes", Attribute.Type),
		edge.To("attribute_options", AttributeOption.Type),
		edge.To("attribute_option_values", AttributeOptionValue.Type),

		// O2M thực: User -> DepartmentMember
		edge.To("dept_memberships", DepartmentMember.Type),
	}
}
