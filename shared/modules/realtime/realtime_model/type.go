package realtime_model

import "encoding/json"

type RealtimeRequest struct {
	UserID  int              `json:"user_id"`
	Message RealtimeEnvelope `json:"message"`
}

type RealtimeAllRequest struct {
	Message RealtimeEnvelope `json:"message"`
}

type RealtimeEnvelope struct {
	Type    string          `json:"type"`    // e.g. "order_created"
	Payload json.RawMessage `json:"payload"` // serialized JSON of payload
}

// Interface gợi ý nếu cần xử lý trước khi gửi
type RealtimePayload interface {
	EventType() string
}
