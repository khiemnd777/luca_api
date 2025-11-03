package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/auditlog/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type AuditLogHandler struct {
	svc *service.AuditLogService
}

func NewAuditLogHandler(svc *service.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{svc: svc}
}

func (h *AuditLogHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/logs", h.Create)
	app.RouterGet(router, "", h.ListPaginated)
}

func (h *AuditLogHandler) Create(c *fiber.Ctx) error {
	type body struct {
		Action   string         `json:"action"`
		Module   string         `json:"module"`
		TargetID int            `json:"target_id"`
		Data     map[string]any `json:"extra_data"`
	}
	var b body
	if err := c.BodyParser(&b); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	userID, _ := utils.GetUserIDInt(c)
	if err := h.svc.Log(c.UserContext(), userID, b.Action, b.Module, b.TargetID, b.Data); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusCreated)
}

func (h *AuditLogHandler) ListPaginated(c *fiber.Ctx) error {
	module := utils.GetQueryAsString(c, "module")
	targetID := utils.GetQueryAsInt(c, "target_id")
	page := utils.GetQueryAsInt(c, "page")
	limit := utils.GetQueryAsInt(c, "limit", 10)

	items, hasMore, err := h.svc.ListByTargetPaginated(c.UserContext(), module, targetID, limit, page)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "Failed to list audit logs")
	}

	return c.JSON(fiber.Map{
		"data":     items,
		"has_more": hasMore,
	})
}
