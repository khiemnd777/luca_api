package client_error

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	Code    int    `json:"statusCode"`
	Message string `json:"statusMessage"`
}

func ResponseError(c *fiber.Ctx, statusCode int, err error, extraMessage ...string) error {
	message := "Server error"
	if len(extraMessage) > 0 && extraMessage[0] != "" {
		message = fmt.Sprintf("%s: %s", message, extraMessage[0])
	}
	if os.Getenv("APP_ENV") == "development" {
		if err != nil {
			message = fmt.Sprintf("%s\n%s", message, err.Error())
		}
	}
	errResp := ErrorResponse{
		Code:    statusCode,
		Message: message,
	}
	return c.Status(statusCode).JSON(errResp)
}

type UnexpectedResponse struct {
	Code    int    `json:"statusCode"`
	Message string `json:"statusMessage"`
}

func ResponseServiceMessage(c *fiber.Ctx, statusCode int, extraMessage ...string) error {
	message := "Service message"
	if len(extraMessage) > 0 && extraMessage[0] != "" {
		message = extraMessage[0]
	}
	errResp := UnexpectedResponse{
		Code:    statusCode,
		Message: message,
	}
	return c.Status(fiber.StatusOK).JSON(errResp)
}
