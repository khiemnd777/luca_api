package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/circuitbreaker"
	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/logger"
)

// RetryOptions configures retry behavior for a route
type RetryOptions struct {
	MaxAttempts int
	Delay       time.Duration
	ShouldRetry func(error) bool
}

func isWebSocketRequest(c *fiber.Ctx) bool {
	// RFC 6455
	if c.Method() != fiber.MethodGet {
		return false
	}
	if strings.ToLower(c.Get("Upgrade")) != "websocket" {
		return false
	}
	if !strings.Contains(strings.ToLower(c.Get("Connection")), "upgrade") {
		return false
	}
	return true
}

// WrapHandler applies Circuit Breaker + Retry logic to a single handler
func WrapHandler(name string, h fiber.Handler, opts ...RetryOptions) fiber.Handler {
	cfgRetry := config.Get().Retry

	// default retry config
	defaultRetry := RetryOptions{
		MaxAttempts: cfgRetry.MaxAttempts,
		Delay:       cfgRetry.Delay,
		ShouldRetry: func(err error) bool {
			if ferr, ok := err.(*fiber.Error); ok && ferr.Code >= 400 && ferr.Code < 500 {
				return false
			}
			return err != nil
		},
	}
	// fallback to default
	retry := defaultRetry
	if len(opts) > 0 {
		retry = opts[0]
	}

	return func(c *fiber.Ctx) error {
		if isWebSocketRequest(c) {
			logger.Info("🔌 WS bypass circuit: " + name)
			return h(c)
		}

		var err error
		for i := 0; i < retry.MaxAttempts; i++ {
			_, err = circuitbreaker.Run(name, func(ctx context.Context) (interface{}, error) {
				handleErr := h(c)

				if ferr, ok := handleErr.(*fiber.Error); ok && ferr.Code >= 400 && ferr.Code < 500 {
					return nil, nil
				}

				return nil, handleErr
			})

			if err == nil || !retry.ShouldRetry(err) {
				return err
			}

			logger.Warn(fmt.Sprintf("🔁 Retry [%s] #%d failed: %v", name, i+1, err))
			time.Sleep(retry.Delay)
		}

		log.Printf("❌ Handler failed after retries: %s", name)
		return err
	}
}

// WrapHandlers applies middleware(s) and wraps the final handler with CB + Retry
func WrapHandlers(name string, handlers []fiber.Handler, opts ...RetryOptions) []fiber.Handler {
	if len(handlers) == 0 {
		panic("no handlers provided")
	}
	last := handlers[len(handlers)-1]
	middlewares := handlers[:len(handlers)-1]
	wrapped := WrapHandler(name, last, opts...)
	return append(middlewares, wrapped)
}
