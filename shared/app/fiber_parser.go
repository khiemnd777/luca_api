package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
)

func ParseBody[T any](c *fiber.Ctx) (*T, error) {
	var body T
	if err := c.BodyParser(&body); err != nil {
		return nil, client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request body")
	}
	return &body, nil
}
