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
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type OrderItemMaterialHandler struct {
	svc  service.OrderItemMaterialService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewOrderItemMaterialHandler(svc service.OrderItemMaterialService, deps *module.ModuleDeps[config.ModuleConfig]) *OrderItemMaterialHandler {
	return &OrderItemMaterialHandler{svc: svc, deps: deps}
}

func (h *OrderItemMaterialHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/order/item/material/loaner/list", h.GetLoanerMaterials)
}

func (h *OrderItemMaterialHandler) GetLoanerMaterials(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := table.ParseTableQuery(c, 20)
	res, err := h.svc.GetLoanerMaterials(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}
