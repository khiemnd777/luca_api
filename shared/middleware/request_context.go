package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func AttachRequestContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := strings.TrimSpace(c.Get("X-Request-ID"))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set("X-Request-ID", requestID)
		c.Locals("requestID", requestID)

		ctx := logger.ContextWithFields(c.UserContext(), map[string]any{
			"request_id": requestID,
			"method":     c.Method(),
			"path":       c.Path(),
		})
		c.SetUserContext(ctx)

		return c.Next()
	}
}
