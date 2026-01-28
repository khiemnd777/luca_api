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
	"github.com/khiemnd777/andy_api/shared/utils/table"
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
	app.RouterGet(router, "/:dept_id<int>/staff/:staff_id/order/processes/in-progresses", h.GetInProgressesByAssignedID)
	app.RouterGet(router, "/:dept_id<int>/order/:order_id<int>/historical/:order_item_id<int>/processes", h.Processes)
	app.RouterGet(router, "/:dept_id<int>/order/:order_id<int>/historical/:order_item_id<int>/processes/in-progresses", h.GetInProgressesByOrderItemID)
	app.RouterGet(router, "/:dept_id<int>/order/processes/in-progress/:in_progress_id<int>", h.GetInProgressByID)
	app.RouterGet(router, "/:dept_id<int>/order/processes/:process_id<int>/in-progresses", h.GetInProgressesByProcessID)
	app.RouterGet(router, "/:dept_id<int>/order/:order_id<int>/historical/:order_item_id<int>/processes/check-out/latest", h.GetCheckoutLatest)
	app.RouterGet(router, "/:dept_id<int>/order/:order_id<int>/historical/:order_item_id<int>/processes/check-in-out/prepare", h.PrepareCheckInOrOut)
	app.RouterGet(router, "/:dept_id<int>/order/processes/check-in-out/prepare-by-code", h.PrepareCheckInOrOutByCode)
	app.RouterPost(router, "/:dept_id<int>/order/:order_id<int>/historical/:order_item_id<int>/processes/check-in-out", h.CheckInOrOut)
	app.RouterPost(router, "/:dept_id<int>/order/processes/in-progress/:in_progress_id<int>/assign", h.Assign)
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

func (h *OrderItemProcessHandler) GetInProgressByID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	inProgressID, _ := utils.GetParamAsInt(c, "in_progress_id")
	if inProgressID <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}

	dto, err := h.svc.GetInProgressByID(c.UserContext(), int64(inProgressID))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *OrderItemProcessHandler) GetInProgressesByProcessID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	processID, _ := utils.GetParamAsInt(c, "process_id")
	if processID <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}

	res, err := h.svc.GetInProgressesByProcessID(c.UserContext(), int64(processID))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *OrderItemProcessHandler) GetInProgressesByOrderItemID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	orderID, orderItemID, err := h.parseOrderParams(c)
	if err != nil {
		return err
	}

	res, err := h.svc.GetInProgressesByOrderItemID(c.UserContext(), int64(orderID), int64(orderItemID))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *OrderItemProcessHandler) GetCheckoutLatest(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	_, orderItemID, err := h.parseOrderParams(c)
	if err != nil {
		return err
	}

	dto, svcErr := h.svc.GetCheckoutLatest(c.UserContext(), int64(orderItemID))
	if svcErr != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, svcErr, svcErr.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *OrderItemProcessHandler) PrepareCheckInOrOut(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.development"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	orderID, orderItemID, err := h.parseOrderParams(c)
	if err != nil {
		return err
	}

	dto, svcErr := h.svc.PrepareCheckInOrOut(c.UserContext(), int64(orderID), int64(orderItemID))
	if svcErr != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, svcErr, svcErr.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *OrderItemProcessHandler) PrepareCheckInOrOutByCode(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.development"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	code := utils.GetQueryAsString(c, "code")
	if code == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid code")
	}

	dto, svcErr := h.svc.PrepareCheckInOrOutByCode(c.UserContext(), code)
	if svcErr != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, svcErr, svcErr.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *OrderItemProcessHandler) CheckInOrOut(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.development"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	orderID, orderItemID, err := h.parseOrderParams(c)
	if err != nil {
		return err
	}

	var (
		checkInOrOutData *model.OrderItemProcessInProgressDTO
	)
	if len(c.Body()) > 0 {
		payload, err := app.ParseBody[model.OrderItemProcessInProgressDTO](c)
		if err != nil {
			return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
		}
		checkInOrOutData = payload
	}

	if checkInOrOutData == nil {
		checkInOrOutData, err = h.svc.PrepareCheckInOrOut(c.UserContext(), int64(orderID), int64(orderItemID))
		if err != nil {
			return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
		}
	} else {
		if checkInOrOutData.OrderItemID == 0 {
			checkInOrOutData.OrderItemID = int64(orderItemID)
		}
		if checkInOrOutData.OrderID == nil {
			oid := int64(orderID)
			checkInOrOutData.OrderID = &oid
		}
	}

	userID, _ := utils.GetUserIDInt(c)
	deptID, _ := utils.GetDeptIDInt(c)

	dto, err := h.svc.CheckInOrOut(c.UserContext(), deptID, userID, checkInOrOutData)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *OrderItemProcessHandler) Assign(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.development"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	inProgressID, _ := utils.GetParamAsInt(c, "in_progress_id")
	if inProgressID <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}

	payload, err := app.ParseBody[struct {
		AssignedID   *int64  `json:"assigned_id"`
		AssignedName *string `json:"assigned_name"`
		Note         *string `json:"note"`
	}](c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}

	dto, err := h.svc.Assign(c.UserContext(), int64(inProgressID), payload.AssignedID, payload.AssignedName, payload.Note)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
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

func (h *OrderItemProcessHandler) GetInProgressesByAssignedID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.development"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	staffID, _ := utils.GetParamAsInt(c, "staff_id")
	q := table.ParseTableQuery(c, 20)
	res, err := h.svc.GetInProgressesByAssignedID(c.UserContext(), int64(staffID), q)
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
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
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

func (h *OrderItemProcessHandler) parseOrderParams(c *fiber.Ctx) (int, int, error) {
	orderID, _ := utils.GetParamAsInt(c, "order_id")
	if orderID <= 0 {
		return 0, 0, client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order id")
	}

	orderItemID, _ := utils.GetParamAsInt(c, "order_item_id")
	if orderItemID <= 0 {
		return 0, 0, client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order item id")
	}

	return orderID, orderItemID, nil
}
