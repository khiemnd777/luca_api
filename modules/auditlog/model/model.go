package model

import (
	"time"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

type AuditLogModel struct {
	ID        int             `json:"id"`
	UserID    int             `json:"user_id"`
	Action    string          `json:"action"`
	Module    string          `json:"module"`
	TargetID  *int            `json:"target_id"`
	Data      map[string]any  `json:"data"`
	User      *generated.User `json:"user,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}
