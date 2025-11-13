package service

import (
	"context"

	"github.com/khiemnd777/andy_api/shared/modules/search/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

func (s *searchService) InitPubSubEvents() {
	pubsub.SubscribeAsync("search:upsert", func(payload *model.Doc) error {
		ctx := context.Background()
		return s.Upsert(ctx, *payload)
	})
}
