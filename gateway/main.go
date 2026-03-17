package main

import (
	"github.com/gofiber/fiber/v2"
	gateway "github.com/khiemnd777/andy_api/gateway/runtime"
	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func main() {
	logger.Init()
	logger.SetComponent("gateway")

	config.Init(utils.GetFullPath("config.yaml"))
	logger.Configure(logger.Options{
		ServiceName:  config.Get().Observability.ServiceName,
		Environment:  config.Get().Observability.Environment,
		Level:        config.Get().Observability.Logs.Level,
		RedactFields: config.Get().Observability.Logs.RedactFields,
		Component:    "gateway",
	})

	logger.Info("Starting API Gateway...")
	app := fiber.New()

	if err := gateway.Start(app); err != nil {
		logger.Error("Gateway error", err)
	}
}
