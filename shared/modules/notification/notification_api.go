package notification

import (
	"github.com/google/uuid"
	"github.com/khiemnd777/andy_api/shared/modules/notification/model"
	"github.com/khiemnd777/andy_api/shared/modules/realtime"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

// Example:
//
//	notification.Notify(c.UserContext(), userID, notifierID, "ws:test", map[string]any{
//		"message": "Andy xin chào!",
//		"time":    time.Now().Format(time.RFC3339),
//	})
func Notify(receiverID, notifierID int, notificationType string, data map[string]any) {
	messageID := uuid.NewString()

	if data != nil {
		data["message_id"] = messageID
	}

	pubsub.PublishAsync("notification:notify", model.NotifyRequest{
		Type:       notificationType,
		UserID:     receiverID,
		NotifierID: notifierID,
		MessageID:  messageID,
		Data:       data,
	})

	realtime.Send(receiverID, notificationType, data)
}
