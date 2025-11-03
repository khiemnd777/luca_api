package schema

import (
	"errors"
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
		field.String("type").NotEmpty().
			Validate(func(s string) error {
				switch s {
				case "order_request:new":
					return nil
				case "order_request:accepted":
					return nil
				case "subscriber:new_product":
					return nil
				case "subscriber:updated_product":
					return nil
				case "credit:spend":
					return nil
				case "checkout:transfer":
					return nil
				case "credit:reserved":
					return nil
				case "credit:activated":
					return nil
				case "order:message":
					return nil
				case "order_request:message":
					return nil
				default:
					return errors.New("invalid notification type")
				}
			}),
		field.Bool("read").Default(false),
		field.JSON("data", map[string]any{}).Optional(),
		field.Bool("deleted").Default(false),
	}
}
