package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ProductBrandName struct {
	ent.Schema
}

func (ProductBrandName) Fields() []ent.Field {
	return []ent.Field{
		field.Int("product_id"),
		field.Int("brand_name_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (ProductBrandName) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("brand_names").
			Field("product_id").
			Unique().
			Required(),

		edge.From("brand_name", BrandName.Type).
			Ref("products").
			Field("brand_name_id").
			Unique().
			Required(),
	}
}

func (ProductBrandName) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id", "brand_name_id").Unique(),
		index.Fields("brand_name_id"),
		index.Fields("product_id"),
	}
}
