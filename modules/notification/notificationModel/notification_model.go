package notificationModel

import (
	"time"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

type Notification struct {
	ID         int             `json:"id,omitempty"`
	UserID     int             `json:"user_id,omitempty"`
	NotifierID int             `json:"notifier_id,omitempty"`
	CreatedAt  time.Time       `json:"created_at,omitempty"`
	Type       string          `json:"type,omitempty"`
	Read       bool            `json:"read,omitempty"`
	Data       map[string]any  `json:"data,omitempty"`
	Notifier   *generated.User `json:"notifier,omitempty"`
}
