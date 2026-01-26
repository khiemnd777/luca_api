package service

import (
	"context"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

func (s *caseDailyStatsService) InitPubSubEvents() {
	pubsub.SubscribeAsync("order:completed", func(payload *model.CaseDailyStatsUpsert) error {
		ctx := context.Background()
		turnaroundsec := payload.CompletedAt.Sub(payload.ReceivedAt).Seconds()
		return s.UpsertOne(ctx, payload.CompletedAt, payload.DepartmentID, int64(turnaroundsec))
	})
}
