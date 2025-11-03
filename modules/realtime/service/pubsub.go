package service

import (
	"encoding/json"

	"github.com/khiemnd777/andy_api/shared/modules/realtime/realtime_model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

func (s *Hub) InitPubSubEvents() {
	pubsub.Subscribe("realtime:send", func(payload *realtime_model.RealtimeRequest) error {
		msg, _ := json.Marshal(payload.Message)
		s.SendToUser(payload.UserID, msg)
		return nil
	})
}
