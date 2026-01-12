package service

import (
	"encoding/json"

	"github.com/khiemnd777/andy_api/shared/modules/realtime/realtime_model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

func (s *Hub) InitPubSubEvents() {
	pubsub.Subscribe("realtime:send", func(payload *realtime_model.RealtimeRequest) error {
		msg, _ := json.Marshal(payload.Message)
		s.BroadcastTo(payload.UserID, msg)
		return nil
	})

	pubsub.Subscribe("realtime:broadcast:all", func(payload *realtime_model.RealtimeAllRequest) error {
		msg, _ := json.Marshal(payload.Message)
		s.BroadcastAll(msg)
		return nil
	})
}
