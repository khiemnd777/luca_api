package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/main/config"
	orderservice "github.com/khiemnd777/andy_api/modules/main/features/order/service"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/module"
)

var deliveryQRStartRateLimitStore = struct {
	mu      sync.Mutex
	windows map[string][]time.Time
}{
	windows: make(map[string][]time.Time),
}

func DeliverySessionAuthMiddleware(deps *module.ModuleDeps[config.ModuleConfig]) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := strings.TrimSpace(c.Cookies(orderservice.DeliveryQRSessionCookieName))
		if sessionID == "" {
			return c.Next()
		}

		session, err := orderservice.LoadDeliveryQRSession(sessionID)
		if err != nil {
			logger.Warn("delivery_qr_session_invalid", "session_id", sessionID, "ip", c.IP(), "error", err.Error())
			return c.Next()
		}
		if session == nil || time.Now().After(session.ExpiresAt) {
			logger.Warn("delivery_qr_session_invalid", "session_id", sessionID, "ip", c.IP(), "reason", "expired_or_missing")
			return c.Next()
		}

		perms := []string{"order.delivery"}
		permSet := map[string]struct{}{
			"order.delivery": {},
		}

		c.Locals("permissions", perms)
		c.Locals("permSet", permSet)
		c.Locals("delivery_session", *session)
		c.Locals("auth_type", "delivery_qr")

		return c.Next()
	}
}

func DeliveryQRStartRateLimitMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()
		cutoff := now.Add(-time.Minute)

		deliveryQRStartRateLimitStore.mu.Lock()
		hits := deliveryQRStartRateLimitStore.windows[ip]
		filtered := hits[:0]
		for _, hitAt := range hits {
			if hitAt.After(cutoff) {
				filtered = append(filtered, hitAt)
			}
		}
		if len(filtered) >= 10 {
			deliveryQRStartRateLimitStore.windows[ip] = filtered
			deliveryQRStartRateLimitStore.mu.Unlock()
			logger.Warn("delivery_qr_session_invalid", "ip", c.IP(), "reason", "rate_limit_exceeded")
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"statusCode":    fiber.StatusTooManyRequests,
				"statusMessage": "Too many QR start requests",
			})
		}
		deliveryQRStartRateLimitStore.windows[ip] = append(filtered, now)
		deliveryQRStartRateLimitStore.mu.Unlock()

		return c.Next()
	}
}
