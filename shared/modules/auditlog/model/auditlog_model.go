package auditlogmodel

type AuditLogRequest struct {
	UserID   int            `json:"user_id"`
	Action   string         `json:"action"`
	Module   string         `json:"module"`
	TargetID int64          `json:"target_id"`
	Data     map[string]any `json:"extra_data"`
}
