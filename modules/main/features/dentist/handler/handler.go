package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/dentist/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type DentistHandler struct {
	svc  service.DentistService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewDentistHandler(svc service.DentistService, deps *module.ModuleDeps[config.ModuleConfig]) *DentistHandler {
	return &DentistHandler{svc: svc, deps: deps}
}

func (h *DentistHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/dentist/list", h.List)
	app.RouterGet(router, "/:dept_id<int>/dentist/search", h.Search)
	app.RouterGet(router, "/:dept_id<int>/clinic/:clinic_id<int>/dentists", h.ListByClinicID)
	app.RouterGet(router, "/:dept_id<int>/dentist/:id<int>", h.GetByID)
	app.RouterPost(router, "/:dept_id<int>/dentist", h.Create)
	app.RouterPut(router, "/:dept_id<int>/dentist/:id<int>", h.Update)
	app.RouterDelete(router, "/:dept_id<int>/dentist/:id<int>", h.Delete)
}

func (h *DentistHandler) List(c *fiber.Ctx) error {
	q := table.ParseTableQuery(c, 20)
	res, err := h.svc.List(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *DentistHandler) ListByClinicID(c *fiber.Ctx) error {
	q := table.ParseTableQuery(c, 20)
	clinicID, _ := utils.GetParamAsInt(c, "clinic_id")
	res, err := h.svc.ListByClinicID(c.UserContext(), clinicID, q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *DentistHandler) Search(c *fiber.Ctx) error {
	q := dbutils.ParseSearchQuery(c, 20)
	res, err := h.svc.Search(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *DentistHandler) GetByID(c *fiber.Ctx) error {
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

func (h *DentistHandler) Create(c *fiber.Ctx) error {
	var payload model.DentistDTO
	if err := c.BodyParser(&payload); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	if payload.Name == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "name is required")
	}

	dto, err := h.svc.Create(c.UserContext(), payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(dto)
}

func (h *DentistHandler) Update(c *fiber.Ctx) error {
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}

	var payload model.DentistDTO
	if err := c.BodyParser(&payload); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	payload.ID = id

	dto, err := h.svc.Update(c.UserContext(), payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *DentistHandler) Delete(c *fiber.Ctx) error {
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
