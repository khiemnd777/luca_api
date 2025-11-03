package middleware

import (
	"github.com/gofiber/fiber/v2"
	audit "github.com/khiemnd777/andy_api/modules/auditlog/worker"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/worker"
)

type AuditLogConfig struct {
	Action    string
	Module    string
	ExtractID func(c *fiber.Ctx) *int
	Extra     func(c *fiber.Ctx) map[string]any
}

func AuditLog(cfg AuditLogConfig) fiber.Handler {

	return func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			return err
		}
		if c.Response().StatusCode() >= 300 {
			return nil
		}

		_, ok := utils.GetUserIDInt(c)
		if !ok {
			return nil
		}

		targetID := cfg.ExtractID(c)
		extra := cfg.Extra(c)

		log := audit.LogRequest{
			Action:    cfg.Action,
			Module:    cfg.Module,
			TargetID:  targetID,
			ExtraData: extra,
		}

		worker.Enqueue("audit", log)

		return nil
	}
}
