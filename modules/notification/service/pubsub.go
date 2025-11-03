package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/modules/api/notification_model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

func (s *NotificationService) InitPubSubEvents() {
	pubsub.SubscribeAsync("notification:notify", func(payload *notification_model.NotifyRequest) error {
		ctx := context.Background()
		if _, err := s.Create(ctx, payload.MessageID, payload.UserID, payload.NotifierID, payload.Type, payload.Data); err != nil {
			logger.Error(fmt.Sprintf("❌ Failed to create notification: %v", err))
		}
		return nil
	})
}
