package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/department/model"
	"github.com/khiemnd777/andy_api/modules/main/department/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type DepartmentHandler struct {
	svc  service.DepartmentService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewDepartmentHandler(svc service.DepartmentService, deps *module.ModuleDeps[config.ModuleConfig]) *DepartmentHandler {
	return &DepartmentHandler{svc: svc, deps: deps}
}
func (h *DepartmentHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>", h.List)
	app.RouterGet(router, "/:dept_id<int>/child/:child_dept_id<int>", h.GetByID)
	app.RouterGet(router, "/:dept_id<int>/children", h.ChildrenList)
	app.RouterPost(router, "/:dept_id<int>/child/:child_dept_id<int>", h.Create)
	app.RouterPut(router, "/:dept_id<int>/child/:child_dept_id<int>", h.Update)
	app.RouterDelete(router, "/:dept_id<int>/child/:child_dept_id<int>", h.Delete)
	app.RouterGet(router, "/me", h.MyFirstDepartment)
}

func (h *DepartmentHandler) List(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "department.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := table.ParseTableQuery(c, table.DefaultLimit)

	res, err := h.svc.List(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *DepartmentHandler) GetByID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "department.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, err := strconv.Atoi(c.Params("child_dept_id"))
	if err != nil || id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}
	res, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *DepartmentHandler) GetBySlug(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "department.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	slug := c.Params("slug")
	res, err := h.svc.GetBySlug(c.UserContext(), slug)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *DepartmentHandler) MyFirstDepartment(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "department.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	userID, _ := utils.GetUserIDInt(c)
	res, err := h.svc.GetFirstDepartmentOfUser(c.UserContext(), userID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *DepartmentHandler) ChildrenList(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "department.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	parentID, err := strconv.Atoi(c.Params("dept_id"))
	if err != nil || parentID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}
	q := table.ParseTableQuery(c, table.DefaultLimit)

	res, err := h.svc.ChildrenList(c.UserContext(), parentID, q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *DepartmentHandler) Create(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "department.create"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	var in model.DepartmentDTO
	if err := c.BodyParser(&in); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	if in.Name == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "name is required")
	}
	res, err := h.svc.Create(c.UserContext(), in)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *DepartmentHandler) Update(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "department.update"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	id, err := strconv.Atoi(c.Params("child_dept_id"))
	if err != nil || id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}

	var in model.DepartmentDTO
	if err := c.BodyParser(&in); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	in.ID = id
	if in.Name == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "name is required")
	}
	userID, _ := utils.GetUserIDInt(c)
	res, err := h.svc.Update(c.UserContext(), in, userID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *DepartmentHandler) Delete(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "department.delete"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, err := strconv.Atoi(c.Params("child_dept_id"))
	if err != nil || id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
	})
}
