package push_notification_model

type PushTokenRequest struct {
	DeviceToken string         `json:"device_token" validate:"required"`
	Title       string         `json:"title,omitempty"`
	Message     string         `json:"message,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

type PushUserRequest struct {
	UserID  int            `json:"user_id" validate:"required"`
	Title   string         `json:"title,omitempty"`
	Message string         `json:"message,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}
