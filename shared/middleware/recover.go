package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func RecoverAndWrapError() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic: %v", r)
				logger.Error("Recovered from panic", "error", err)
				_ = client_error.ResponseError(c, fiber.StatusInternalServerError, err)
			}
		}()
		return c.Next()
	}
}
