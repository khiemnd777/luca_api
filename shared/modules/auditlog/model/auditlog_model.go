package auditlog_model

type AuditLogRequest struct {
	UserID   int            `json:"user_id"`
	Action   string         `json:"action"`
	Module   string         `json:"module"`
	TargetID int            `json:"target_id"`
	Data     map[string]any `json:"extra_data"`
}
