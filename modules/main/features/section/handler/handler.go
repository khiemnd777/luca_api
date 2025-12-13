package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/section/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type SectionHandler struct {
	svc  service.SectionService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewSectionHandler(svc service.SectionService, deps *module.ModuleDeps[config.ModuleConfig]) *SectionHandler {
	return &SectionHandler{svc: svc, deps: deps}
}

func (h *SectionHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/section/list", h.List)
	app.RouterGet(router, "/:dept_id<int>/section/search", h.Search)
	app.RouterGet(router, "/:dept_id<int>/staff/:staff_id<int>/sections", h.ListBySectionID)
	app.RouterGet(router, "/:dept_id<int>/section/:id<int>", h.GetByID)
	app.RouterPost(router, "/:dept_id<int>/section", h.Create)
	app.RouterPut(router, "/:dept_id<int>/section/:id<int>", h.Update)
	app.RouterDelete(router, "/:dept_id<int>/section/:id<int>", h.Delete)
}

func (h *SectionHandler) List(c *fiber.Ctx) error {
	q := table.ParseTableQuery(c, 20)
	deptID, _ := utils.GetDeptIDInt(c)
	res, err := h.svc.List(c.UserContext(), deptID, q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *SectionHandler) ListBySectionID(c *fiber.Ctx) error {
	q := table.ParseTableQuery(c, 20)
	staffID, _ := utils.GetParamAsInt(c, "staff_id")
	res, err := h.svc.ListByStaffID(c.UserContext(), staffID, q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *SectionHandler) Search(c *fiber.Ctx) error {
	q := dbutils.ParseSearchQuery(c, 20)
	deptID, _ := utils.GetDeptIDInt(c)
	res, err := h.svc.Search(c.UserContext(), deptID, q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *SectionHandler) GetByID(c *fiber.Ctx) error {
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}

	dto, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *SectionHandler) Create(c *fiber.Ctx) error {
	var payload model.SectionDTO
	if err := c.BodyParser(&payload); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	if payload.Name == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "name is required")
	}

	deptID, _ := utils.GetDeptIDInt(c)
	payload.DepartmentID = deptID

	dto, err := h.svc.Create(c.UserContext(), payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(dto)
}

func (h *SectionHandler) Update(c *fiber.Ctx) error {
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}

	var payload model.SectionDTO
	if err := c.BodyParser(&payload); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	payload.ID = id
	deptID, _ := utils.GetDeptIDInt(c)
	payload.DepartmentID = deptID

	dto, err := h.svc.Update(c.UserContext(), payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *SectionHandler) Delete(c *fiber.Ctx) error {
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
