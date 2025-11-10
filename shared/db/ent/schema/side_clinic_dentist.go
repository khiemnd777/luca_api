package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ClinicDentist struct {
	ent.Schema
}

func (ClinicDentist) Fields() []ent.Field {
	return []ent.Field{
		field.Int("dentist_id"),
		field.Int("clinic_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (ClinicDentist) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("dentist", Dentist.Type).
			Ref("clinics").
			Field("dentist_id").
			Unique().
			Required(),

		edge.From("clinic", Clinic.Type).
			Ref("dentists").
			Field("clinic_id").
			Unique().
			Required(),
	}
}

func (ClinicDentist) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("dentist_id", "clinic_id").Unique(),
		index.Fields("clinic_id"),
		index.Fields("dentist_id"),
		index.Fields("dentist_id", "created_at"),
	}
}
