package handler

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/department/model"
	"github.com/khiemnd777/andy_api/modules/main/department/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
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
	app.RouterGet(router, "/:dept_id<int>", h.GetByID)
	app.RouterGet(router, "/:dept_id<int>/children", h.ChildrenList)
	app.RouterPost(router, "/:dept_id<int>", h.Create)
	app.RouterPut(router, "/:dept_id<int>", h.Update)
	app.RouterDelete(router, "/:dept_id<int>", h.Delete)
	app.RouterGet(router, "/me", h.MyFirstDepartment)
}

func (h *DepartmentHandler) List(c *fiber.Ctx) error {
	limit := parseIntDefault(c.Query("limit"), 50, 1, 200)
	offset := parseIntDefault(c.Query("offset"), 0, 0, 1<<31-1)

	items, total, err := h.svc.List(c.UserContext(), limit, offset)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(fiber.Map{
		"items": items,
		"total": total,
	})
}

func (h *DepartmentHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("dept_id"))
	if err != nil || id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}
	res, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(res)
}

func (h *DepartmentHandler) GetBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	res, err := h.svc.GetBySlug(c.UserContext(), slug)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(res)
}

func (h *DepartmentHandler) MyFirstDepartment(c *fiber.Ctx) error {
	userID, _ := utils.GetUserIDInt(c)
	res, err := h.svc.GetFirstDepartmentOfUser(c.UserContext(), userID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(res)
}

func (h *DepartmentHandler) ChildrenList(c *fiber.Ctx) error {
	logger.Debug("[Here] ChildrenList")
	parentID, err := strconv.Atoi(c.Params("dept_id"))
	if err != nil || parentID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}
	limit := parseIntDefault(c.Query("limit"), 50, 1, 200)
	offset := parseIntDefault(c.Query("offset"), 0, 0, 1<<31-1)

	items, total, err := h.svc.ChildrenList(c.UserContext(), parentID, limit, offset)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(fiber.Map{
		"items": items,
		"total": total,
	})
}

func (h *DepartmentHandler) Create(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "department.manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	var in model.DepartmentDTO
	if err := c.BodyParser(&in); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	if in.Name == "" { // đổi lại theo field thực tế của bạn
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "name is required")
	}
	res, err := h.svc.Create(c.UserContext(), in)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, nil, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *DepartmentHandler) Update(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "department.manage"); err != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	id, err := strconv.Atoi(c.Params("dept_id"))
	if err != nil || id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}

	var in model.DepartmentDTO
	if err := c.BodyParser(&in); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	in.ID = id         // ensure path param wins
	if in.Name == "" { // đổi lại theo field thực tế của bạn
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "name is required")
	}
	userID, _ := utils.GetUserIDInt(c)
	res, err := h.svc.Update(c.UserContext(), in, userID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(res)
}

func (h *DepartmentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("dept_id"))
	if err != nil || id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseIntDefault(s string, def, min, max int) int {
	if v, err := strconv.Atoi(s); err == nil {
		if v < min {
			return min
		}
		if v > max {
			return max
		}
		return v
	}
	return def
}
