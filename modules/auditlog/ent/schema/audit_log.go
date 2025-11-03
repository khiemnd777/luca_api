package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AuditLog holds the schema definition for the AuditLog entity.
type AuditLog struct {
	ent.Schema
}

// Fields of the AuditLog.
func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id").Positive(),
		field.String("action").NotEmpty(),
		field.String("module").NotEmpty(),
		field.Int("target_id").Optional().Nillable(),
		field.JSON("data", map[string]any{}).Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

// Indexes of the AuditLog.
func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("module"),
		index.Fields("created_at"),
	}
}
