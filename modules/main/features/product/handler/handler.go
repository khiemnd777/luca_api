package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/product/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type ProductHandler struct {
	svc  service.ProductService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewProductHandler(svc service.ProductService, deps *module.ModuleDeps[config.ModuleConfig]) *ProductHandler {
	return &ProductHandler{svc: svc, deps: deps}
}

func (h *ProductHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/product/list", h.List)
	app.RouterGet(router, "/:dept_id<int>/product/search", h.Search)
	app.RouterGet(router, "/:dept_id<int>/product/:id<int>", h.GetByID)
	app.RouterPost(router, "/:dept_id<int>/product", h.Create)
	app.RouterPut(router, "/:dept_id<int>/product/:id<int>", h.Update)
	app.RouterDelete(router, "/:dept_id<int>/product/:id<int>", h.Delete)
}

func (h *ProductHandler) List(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "product.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := table.ParseTableQuery(c, 20)
	res, err := h.svc.List(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *ProductHandler) Search(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "product.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := dbutils.ParseSearchQuery(c, 20)
	res, err := h.svc.Search(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *ProductHandler) GetByID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "product.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}

	dto, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *ProductHandler) Create(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "product.create"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	payload, err := app.ParseBody[model.ProductUpsertDTO](c)
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

func (h *ProductHandler) Update(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "product.update"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}

	payload, err := app.ParseBody[model.ProductUpsertDTO](c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	payload.DTO.ID = id

	deptID, _ := utils.GetDeptIDInt(c)

	dto, err := h.svc.Update(c.UserContext(), deptID, payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *ProductHandler) Delete(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "product.delete"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
