package client_error

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type ErrorResponse struct {
	Code    int    `json:"statusCode"`
	Message string `json:"statusMessage"`
}

func callerLocation(skip int) string {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}

	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = fn.Name()
	}

	return fmt.Sprintf("%s:%d (%s)",
		filepath.Base(file),
		line,
		funcName,
	)
}

func ResponseError(c *fiber.Ctx, statusCode int, err error, extraMessage ...string) error {
	message := "Server error"

	if len(extraMessage) > 0 && extraMessage[0] != "" {
		message = fmt.Sprintf("%s: %s", message, extraMessage[0])
	}

	if os.Getenv("APP_ENV") == "development" && err != nil {
		message = fmt.Sprintf("%s\n%s", message, err.Error())
	}

	// location := callerLocation(1)

	// logger.Error(fmt.Sprintf(
	// 	"[ERROR] %s | at %s",
	// 	message,
	// 	location,
	// ))

	logger.Debug(string(debug.Stack()))

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
