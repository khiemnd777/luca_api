package handler

import (
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

type OrderItemProcessHandler struct {
	svc  service.OrderItemProcessService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewOrderItemProcessHandler(svc service.OrderItemProcessService, deps *module.ModuleDeps[config.ModuleConfig]) *OrderItemProcessHandler {
	return &OrderItemProcessHandler{svc: svc, deps: deps}
}

func (h *OrderItemProcessHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/order/:order_id<int>/historical/:order_item_id<int>/processes", h.Processes)
}

func (h *OrderItemProcessHandler) Processes(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	orderID, _ := utils.GetParamAsInt(c, "order_id")
	orderItemID, _ := utils.GetParamAsInt(c, "order_item_id")
	res, err := h.svc.GetByOrderItemID(c.UserContext(), int64(orderID), int64(orderItemID))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}
