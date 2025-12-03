package app

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func ParseBody[T any](c *fiber.Ctx) (*T, error) {
	var body T
	if err := c.BodyParser(&body); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}
	return &body, nil
}
