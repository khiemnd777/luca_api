package notification_model

type NotifyRequest struct {
	UserID     int            `json:"user_id" validate:"required"`
	NotifierID int            `json:"notifier_id" validate:"required"`
	MessageID  string         `json:"message_id"`
	Type       string         `json:"type"`
	Data       map[string]any `json:"data,omitempty"`
}
