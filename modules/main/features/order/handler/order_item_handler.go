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

type OrderItemHandler struct {
	svc  service.OrderItemService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewOrderItemHandler(svc service.OrderItemService, deps *module.ModuleDeps[config.ModuleConfig]) *OrderItemHandler {
	return &OrderItemHandler{svc: svc, deps: deps}
}

func (h *OrderItemHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/order/:order_id<int>/historical/:order_item_id<int>/list", h.Historical)
}

func (h *OrderItemHandler) Historical(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	orderID, _ := utils.GetParamAsInt(c, "order_id")
	orderItemID, _ := utils.GetParamAsInt(c, "order_item_id")
	res, err := h.svc.GetHistoricalByOrderIDAndOrderItemID(c.UserContext(), int64(orderID), int64(orderItemID))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}
