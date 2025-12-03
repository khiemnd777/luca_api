package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
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
	app.RouterGet(router, "/:dept_id<int>/staff/:staff_id/order/processes", h.ProcessesForStaff)
	app.RouterGet(router, "/:dept_id<int>/order/:order_id<int>/historical/:order_item_id<int>/processes", h.Processes)
	app.RouterPut(router, "/:dept_id<int>/order/:order_id<int>/historical/:order_item_id<int>/processes/:order_item_process_id<int>", h.Update)
}

func (h *OrderItemProcessHandler) Processes(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	orderID, _ := utils.GetParamAsInt(c, "order_id")
	orderItemID, _ := utils.GetParamAsInt(c, "order_item_id")
	res, err := h.svc.GetProcessesByOrderItemID(c.UserContext(), int64(orderID), int64(orderItemID))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *OrderItemProcessHandler) ProcessesForStaff(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	staffID, _ := utils.GetParamAsInt(c, "staff_id")
	res, err := h.svc.GetProcessesByAssignedID(c.UserContext(), int64(staffID))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *OrderItemProcessHandler) Update(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.update"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "order_item_process_id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}

	payload, err := app.ParseBody[model.OrderItemProcessUpsertDTO](c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	payload.DTO.ID = int64(id)

	deptID, _ := utils.GetDeptIDInt(c)

	dto, err := h.svc.Update(c.UserContext(), deptID, payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}
