package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ProductRawMaterial struct {
	ent.Schema
}

func (ProductRawMaterial) Fields() []ent.Field {
	return []ent.Field{
		field.Int("product_id"),
		field.Int("raw_material_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (ProductRawMaterial) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("raw_materials").
			Field("product_id").
			Unique().
			Required(),

		edge.From("raw_material", RawMaterial.Type).
			Ref("products").
			Field("raw_material_id").
			Unique().
			Required(),
	}
}

func (ProductRawMaterial) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id", "raw_material_id").Unique(),
		index.Fields("raw_material_id"),
		index.Fields("product_id"),
	}
}
