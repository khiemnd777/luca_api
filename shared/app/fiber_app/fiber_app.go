package fiber_app

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/config"
	"github.com/khiemnd777/andy_api/shared/middleware"
)

type Server struct {
	app *fiber.App
}

func Init() (*Server, *fiber.App) {
	srvCfg := config.Get().Server

	app := fiber.New(fiber.Config{
		BodyLimit: srvCfg.BodyLimitMB * 1024 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if fe, ok := err.(*fiber.Error); ok {
				return client_error.ResponseError(c, fe.Code, fe, fe.Error())
			}

			return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
		},
	})

	// Default health check route
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	app.Use(logger.New())                     // Simple request logger
	app.Use(middleware.RecoverAndWrapError()) // Recover from panics and return error response

	return &Server{app: app}, app
}

func NewFiberApp() *fiber.App {
	srvCfg := config.Get().Server

	app := fiber.New(fiber.Config{
		BodyLimit: srvCfg.BodyLimitMB * 1024 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if fe, ok := err.(*fiber.Error); ok {
				return client_error.ResponseError(c, fe.Code, fe, fe.Message)
			}

			return client_error.ResponseError(c, fiber.StatusInternalServerError, err)
		},
	})
	app.Use(logger.New())                     // Simple request logger
	app.Use(middleware.RecoverAndWrapError()) // Recover from panics and return error response
	return app
}

func (s *Server) Start() {
	addr := fmt.Sprintf("%s:%d", config.Get().Server.Host, config.Get().Server.Port)
	info := fmt.Sprintf("✅ Fiber listening on http://%s", addr)
	log.Println(info)

	if err := s.app.Listen(addr); err != nil {
		log.Fatal(info)
	}
}
