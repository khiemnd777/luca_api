package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/order/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type OrderHandler struct {
	svc  service.OrderService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewOrderHandler(svc service.OrderService, deps *module.ModuleDeps[config.ModuleConfig]) *OrderHandler {
	return &OrderHandler{svc: svc, deps: deps}
}

func (h *OrderHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/order/list", h.List)
	app.RouterGet(router, "/:dept_id<int>/order/in-progress/list", h.InProgressList)
	app.RouterGet(router, "/:dept_id<int>/order/newest/list", h.NewestList)
	app.RouterGet(router, "/:dept_id<int>/order/search", h.Search)
	app.RouterGet(router, "/:dept_id<int>/order/:id<int>", h.GetByID)
	app.RouterGet(router, "/:dept_id<int>/order/:order_id<int>/remake/prepare", h.PrepareForRemakeByOrderID)
	app.RouterGet(router, "/:dept_id<int>/order/:order_id<int>/historical/:order_item_id<int>", h.GetByOrderIDAndOrderItemID)
	app.RouterGet(router, "/:dept_id<int>/order/:order_id<int>/products", h.GetAllOrderProducts)
	app.RouterGet(router, "/:dept_id<int>/order/:order_id<int>/materials", h.GetAllOrderMaterials)
	app.RouterGet(router, "/:dept_id<int>/order/:id<int>/sync-price", h.SyncPrice)
	app.RouterPost(router, "/:dept_id<int>/order", h.Create)
	app.RouterPut(router, "/:dept_id<int>/order/:id<int>", h.Update)
	app.RouterPut(router, "/:dept_id<int>/order/:id<int>/process/:order_item_process_id<int>/change-status/:status", h.UpdateStatus)
	app.RouterDelete(router, "/:dept_id<int>/order/:id<int>", h.Delete)
}

func (h *OrderHandler) List(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := table.ParseTableQuery(c, 20)
	res, err := h.svc.List(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *OrderHandler) InProgressList(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := table.ParseTableQuery(c, 20)
	res, err := h.svc.InProgressList(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *OrderHandler) NewestList(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := table.ParseTableQuery(c, 20)
	res, err := h.svc.NewestList(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *OrderHandler) Search(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := dbutils.ParseSearchQuery(c, 20)
	res, err := h.svc.Search(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *OrderHandler) GetByID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}

	dto, err := h.svc.GetByID(c.UserContext(), int64(id))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *OrderHandler) GetByOrderIDAndOrderItemID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	orderID, _ := utils.GetParamAsInt(c, "order_id")
	if orderID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order id")
	}

	orderItemID, _ := utils.GetParamAsInt(c, "order_item_id")
	if orderID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order item id")
	}

	dto, err := h.svc.GetByOrderIDAndOrderItemID(c.UserContext(), int64(orderID), int64(orderItemID))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *OrderHandler) PrepareForRemakeByOrderID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	orderID, _ := utils.GetParamAsInt(c, "order_id")
	if orderID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order id")
	}

	dto, err := h.svc.PrepareForRemakeByOrderID(c.UserContext(), int64(orderID))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *OrderHandler) GetAllOrderProducts(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	orderID, _ := utils.GetParamAsInt(c, "order_id")
	if orderID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order id")
	}

	products, err := h.svc.GetAllOrderProducts(c.UserContext(), int64(orderID))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(products)
}

func (h *OrderHandler) GetAllOrderMaterials(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	orderID, _ := utils.GetParamAsInt(c, "order_id")
	if orderID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order id")
	}

	materials, err := h.svc.GetAllOrderMaterials(c.UserContext(), int64(orderID))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(materials)
}

func (h *OrderHandler) SyncPrice(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order id")
	}

	total, err := h.svc.SyncPrice(c.UserContext(), int64(id))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"total_price": total,
	})
}

func (h *OrderHandler) Create(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.create"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	payload, err := app.ParseBody[model.OrderUpsertDTO](c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}

	deptID, _ := utils.GetDeptIDInt(c)

	dto, err := h.svc.Create(c.UserContext(), deptID, payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(dto)
}

func (h *OrderHandler) Update(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.update"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}

	payload, err := app.ParseBody[model.OrderUpsertDTO](c)
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

func (h *OrderHandler) UpdateStatus(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.update"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	oipID, _ := utils.GetParamAsInt(c, "order_item_process_id")
	if oipID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid oip id")
	}

	status := utils.GetParamAsString(c, "status")
	if strings.TrimSpace(status) == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid status")
	}

	dto, err := h.svc.UpdateStatus(c.UserContext(), int64(oipID), status)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *OrderHandler) Delete(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.delete"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}
	if err := h.svc.Delete(c.UserContext(), int64(id)); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
