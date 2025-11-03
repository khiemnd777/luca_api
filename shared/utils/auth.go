package utils

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/shared/config"
)

func GetAuthSecret() string {
	cfg, _ := LoadConfig[config.AppConfig](GetFullPath("config.yaml"))

	if envSecret := os.Getenv("AUTH_SECRET"); envSecret != "" {
		return envSecret
	}
	return cfg.Auth.Secret
}

func GetInternalToken() string {
	cfg, _ := LoadConfig[config.AppConfig](GetFullPath("config.yaml"))

	if envSecret := os.Getenv("INTERNAL_TOKEN"); envSecret != "" {
		return envSecret
	}
	return cfg.Auth.InternalAuthToken
}

// GetAccessToken extracts the Bearer token from the Authorization header
func GetAccessToken(c *fiber.Ctx) string {
	authHeader := c.Get("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}
