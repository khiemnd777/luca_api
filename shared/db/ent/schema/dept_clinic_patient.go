package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ClinicPatient struct {
	ent.Schema
}

func (ClinicPatient) Fields() []ent.Field {
	return []ent.Field{
		field.Int("patient_id"),
		field.Int("clinic_id"),
		field.Time("created_at").Default(time.Now),
	}
}

func (ClinicPatient) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("patient", Patient.Type).
			Ref("clinics").
			Field("patient_id").
			Unique().
			Required(),

		edge.From("clinic", Clinic.Type).
			Ref("patients").
			Field("clinic_id").
			Unique().
			Required(),
	}
}

func (ClinicPatient) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("patient_id", "clinic_id").Unique(),
	}
}
