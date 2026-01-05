package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/order/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

const defaultOrderCodeReservationTTL = 15 * time.Minute

type OrderCodeHandler struct {
	svc  service.OrderCodeService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewOrderCodeHandler(svc service.OrderCodeService, deps *module.ModuleDeps[config.ModuleConfig]) *OrderCodeHandler {
	return &OrderCodeHandler{svc: svc, deps: deps}
}

func (h *OrderCodeHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/:dept_id<int>/order/code/reserve", h.Reserve)
}

func (h *OrderCodeHandler) Reserve(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.create"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	ttlSeconds := utils.GetQueryAsInt(c, "ttl_seconds", int(defaultOrderCodeReservationTTL.Seconds()))
	if ttlSeconds <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid ttl_seconds")
	}

	code, expiresAt, err := h.svc.ReserveOrderCode(
		c.UserContext(),
		time.Now(),
		time.Duration(ttlSeconds)*time.Second,
	)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"order_code": code,
		"expires_at": expiresAt,
	})
}
