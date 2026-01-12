package realtime

import (
	"encoding/json"
	"fmt"

	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/modules/realtime/realtime_model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

func SendTyped(userID int, payload realtime_model.RealtimePayload) {
	Send(userID, payload.EventType(), payload)
}

// Example:
//
//	realtime.Send(c.UserContext(), userID, "ws:test", map[string]any{
//		"message": "Andy xin chào!",
//		"time":    time.Now().Format(time.RFC3339),
//	})
func Send(receiverID int, eventType string, data any) {
	raw, err := json.Marshal(data)
	if err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to marshal realtime payload: %v", err))
		return
	}

	envelope := realtime_model.RealtimeEnvelope{
		Type:    eventType,
		Payload: raw,
	}

	pubsub.Publish("realtime:send", realtime_model.RealtimeRequest{
		UserID:  receiverID,
		Message: envelope,
	})
}

func BroadcastTo(receiverID int, eventType string, data any) {
	raw, err := json.Marshal(data)
	if err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to marshal realtime payload: %v", err))
		return
	}

	envelope := realtime_model.RealtimeEnvelope{
		Type:    eventType,
		Payload: raw,
	}

	pubsub.Publish("realtime:send", realtime_model.RealtimeRequest{
		UserID:  receiverID,
		Message: envelope,
	})
}

func BroadcastAll(eventType string, data any) {
	raw, err := json.Marshal(data)
	if err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to marshal realtime payload: %v", err))
		return
	}

	envelope := realtime_model.RealtimeEnvelope{
		Type:    eventType,
		Payload: raw,
	}

	pubsub.Publish("realtime:broadcast:all", realtime_model.RealtimeAllRequest{
		Message: envelope,
	})
}
