package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type MaterialSupplier struct {
	ent.Schema
}

func (MaterialSupplier) Fields() []ent.Field {
	return []ent.Field{
		field.Int("material_id"),
		field.Int("supplier_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (MaterialSupplier) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("material", Material.Type).
			Ref("suppliers").
			Field("material_id").
			Unique().
			Required(),

		edge.From("supplier", Supplier.Type).
			Ref("materials").
			Field("supplier_id").
			Unique().
			Required(),
	}
}

func (MaterialSupplier) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("material_id", "supplier_id").Unique(),
		index.Fields("supplier_id"),
		index.Fields("material_id"),
	}
}
