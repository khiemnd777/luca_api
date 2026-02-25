package service

import (
	"context"

	auditlog_model "github.com/khiemnd777/andy_api/shared/modules/auditlog/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

/*
e.g.:

	pubsub.PublishAsync("log:create", auditlogmodel.AuditLogRequest{
		UserID:   userID,
		Module:   "product",
		Action:   "checkout",
		TargetID: item.ProductID,
		Data: map[string]any{
			"checkout_id":  session.ID,
			"order_id":     session.OrderID,
			"customer_id":  userID,
			"publisher_id": item.PublisherID,
		},
	})
*/
func (s *AuditLogService) InitPubSubEvents() {
	pubsub.SubscribeAsync("log:create", func(payload *auditlog_model.AuditLogRequest) error {
		ctx := context.Background()
		return s.Log(ctx, payload.UserID, payload.Action, payload.Module, payload.TargetID, payload.Data)
	})
}
