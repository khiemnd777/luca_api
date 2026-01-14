package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Folder holds the schema definition for the Folder entity.
type Notification struct {
	ent.Schema
}

// Fields of the Folder.
func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id"),
		field.Int("notifier_id"),
		field.String("message_id").Unique().Optional(),
		field.Time("created_at").Default(time.Now),
		// field.String("type").NotEmpty().
		// 	Validate(func(s string) error {
		// 		switch s {
		// 		case "order:checkout":
		// 			return nil
		// 		default:
		// 			return errors.New("invalid notification type")
		// 		}
		// 	}),
		field.String("type").NotEmpty(),
		field.Bool("read").Default(false),
		field.JSON("data", map[string]any{}).Optional(),
		field.Bool("deleted").Default(false),
	}
}
