package service

import (
	"context"

	auditlog_model "github.com/khiemnd777/andy_api/shared/modules/auditlog/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

func (s *AuditLogService) InitPubSubEvents() {
	pubsub.SubscribeAsync("log:create", func(payload *auditlog_model.AuditLogRequest) error {
		ctx := context.Background()
		return s.Log(ctx, payload.UserID, payload.Action, payload.Module, payload.TargetID, payload.Data)
	})
}
