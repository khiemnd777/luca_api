package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/redis"
)

type JsonHandler func(msg string) error

var (
	ctx               = context.Background()
	channelHandlers   = make(map[string][]JsonHandler)
	pubsubCancelFuncs = make(map[string]context.CancelFunc)
	messageWgMap      = sync.Map{}
	channelMu         sync.RWMutex
)

type asyncMessage struct {
	ID      string `json:"id"`
	Payload string `json:"payload"` // raw JSON
}

// Publish sends a raw string message to a channel.
func publish(channel string, message string) error {
	rdb := redis.GetInstance("pubsub")
	if rdb == nil {
		return fmt.Errorf("redis instance 'pubsub' not available")
	}
	err := rdb.Publish(ctx, channel, message).Err()
	if err != nil {
		logger.Error("❌ Redis PUBLISH error:", err)
	} else {
		logger.Info("📢 Redis PUBLISH [" + channel + "]: " + message)
	}
	return err
}

// Publish marshals an object into JSON and publishes.
func Publish(channel string, payload any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return publish(channel, string(bytes))
}

func PublishAsync(channel string, payload any) error {
	rawBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msgID := uuid.NewString()
	msg := asyncMessage{
		ID:      msgID,
		Payload: string(rawBytes),
	}
	wrappedBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	channelMu.RLock()
	handlers := append([]JsonHandler(nil), channelHandlers[channel]...)
	channelMu.RUnlock()
	wg := &sync.WaitGroup{}
	wg.Add(len(handlers))
	messageWgMap.Store(msgID, wg)

	err = publish(channel, string(wrappedBytes))
	if err != nil {
		return err
	}

	// Wait for all async handlers done
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Debug("✅ ASYNC handlers done [" + msgID + "]")
	case <-time.After(10 * time.Second):
		logger.Warn("⏰ Timeout waiting for async handlers [" + msgID + "]")
	}

	messageWgMap.Delete(msgID)
	return nil
}

// Subscribe listens for messages from a Redis channel and runs a handler.
func subscribe(channel string, handler JsonHandler) {
	channelMu.Lock()
	if _, exists := channelHandlers[channel]; !exists {
		channelHandlers[channel] = []JsonHandler{}
	}
	channelHandlers[channel] = append(channelHandlers[channel], handler)
	if _, exists := pubsubCancelFuncs[channel]; exists {
		channelMu.Unlock()
		return
	}
	subCtx, cancel := context.WithCancel(ctx)
	pubsubCancelFuncs[channel] = cancel
	channelMu.Unlock()

	rdb := redis.GetInstance("pubsub")
	if rdb == nil {
		logger.Warn("⚠️ Redis instance 'pubsub' not available for channel: " + channel)
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("❌ Redis SUBSCRIBE loop panic [%s]: %v", channel, r))
			}
		}()

		logger.Debug("🔔 Redis SUBSCRIBED to channel: " + channel)
		for {
			sub := rdb.Subscribe(ctx, channel)
			ch := sub.Channel()

			for {
				select {
				case msg, ok := <-ch:
					if !ok || msg == nil {
						logger.Warn("⚠️ Redis subscription channel closed, retrying: " + channel)
						_ = sub.Close()
						time.Sleep(time.Second)
						goto retry
					}

					logger.Debug("📨 Redis RECEIVED [" + msg.Channel + "]: " + msg.Payload)

					channelMu.RLock()
					handlers := append([]JsonHandler(nil), channelHandlers[channel]...)
					channelMu.RUnlock()

					var wg sync.WaitGroup
					for _, h := range handlers {
						wg.Add(1)
						go func(handler JsonHandler) {
							defer wg.Done()
							defer func() {
								if r := recover(); r != nil {
									logger.Error(fmt.Sprintf("❌ Redis handler panic [%s]: %v", channel, r))
								}
							}()
							_ = handler(msg.Payload)
						}(h)
					}
					wg.Wait()
				case <-subCtx.Done():
					_ = sub.Close()
					logger.Debug("🔕 Redis UNSUBSCRIBED from channel: " + channel)
					return
				}
			}
		retry:
		}
	}()
}

// Subscribe subscribes to a channel and decodes each message as JSON into T.
func Subscribe[T any](channel string, handler func(payload *T) error) {
	/* Example:
	type UserUpdatedEvent struct {
		UserID string `json:"userId"`
	}
	pubsub.SubscribeJSON[UserUpdatedEvent]("user:updated", func(event *UserUpdatedEvent) {
		cache.InvalidateKeys("user:" + event.UserID)
	})
	pubsub.PublishJSON("user:updated", UserUpdatedEvent{UserID: "123"})
	*/
	subscribe(channel, func(msg string) error {
		var result T
		if err := json.Unmarshal([]byte(msg), &result); err != nil {
			logger.Warn("❌ Redis JSON Unmarshal failed [" + channel + "]: " + err.Error())
			return err
		}
		return handler(&result)
	})
}

func SubscribeAsync[T any](channel string, handler func(payload *T) error) {
	subscribe(channel, func(msg string) error {
		var wrapper asyncMessage
		if err := json.Unmarshal([]byte(msg), &wrapper); err != nil {
			logger.Warn("❌ Redis asyncMessage unmarshal failed [" + channel + "]: " + err.Error())
			return err
		}

		var result T
		if err := json.Unmarshal([]byte(wrapper.Payload), &result); err != nil {
			logger.Warn("❌ Redis async payload unmarshal failed [" + channel + "]: " + err.Error())
			return err
		}

		defer func() {
			if v, ok := messageWgMap.Load(wrapper.ID); ok {
				v.(*sync.WaitGroup).Done()
			}
		}()

		return handler(&result)
	})
}

func equalHandlers(a, b func(msg string) error) bool {
	return fmt.Sprintf("%p", a) == fmt.Sprintf("%p", b)
}

// Unsubscribe cancels listening to a Redis channel.
func Unsubscribe(channel string, handlerToRemove JsonHandler) {
	channelMu.Lock()
	defer channelMu.Unlock()

	handlers := channelHandlers[channel]
	newHandlers := make([]JsonHandler, 0, len(handlers))

	for _, h := range handlers {
		if !equalHandlers(h, handlerToRemove) {
			newHandlers = append(newHandlers, h)
		}
	}

	if len(newHandlers) == 0 {
		if cancel, ok := pubsubCancelFuncs[channel]; ok {
			cancel()
			delete(pubsubCancelFuncs, channel)
		}
		delete(channelHandlers, channel)
	} else {
		channelHandlers[channel] = newHandlers
	}
}
