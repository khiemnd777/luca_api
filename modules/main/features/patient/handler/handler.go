package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/patient/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type PatientHandler struct {
	svc  service.PatientService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewPatientHandler(svc service.PatientService, deps *module.ModuleDeps[config.ModuleConfig]) *PatientHandler {
	return &PatientHandler{svc: svc, deps: deps}
}

func (h *PatientHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/patient/list", h.List)
	app.RouterGet(router, "/:dept_id<int>/patient/search", h.Search)
	app.RouterGet(router, "/:dept_id<int>/clinic/:clinic_id<int>/patients", h.ListByClinicID)
	app.RouterGet(router, "/:dept_id<int>/patient/:id<int>", h.GetByID)
	app.RouterPost(router, "/:dept_id<int>/patient", h.Create)
	app.RouterPut(router, "/:dept_id<int>/patient/:id<int>", h.Update)
	app.RouterDelete(router, "/:dept_id<int>/patient/:id<int>", h.Delete)
}

func (h *PatientHandler) List(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "clinic.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := table.ParseTableQuery(c, 20)
	res, err := h.svc.List(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *PatientHandler) ListByClinicID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "clinic.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := table.ParseTableQuery(c, 20)
	clinicID, _ := utils.GetParamAsInt(c, "clinic_id")
	res, err := h.svc.ListByClinicID(c.UserContext(), clinicID, q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *PatientHandler) Search(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "clinic.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := dbutils.ParseSearchQuery(c, 20)
	res, err := h.svc.Search(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *PatientHandler) GetByID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "clinic.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
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

func (h *PatientHandler) Create(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "clinic.create"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	var payload model.PatientDTO
	if err := c.BodyParser(&payload); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	if payload.Name == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "name is required")
	}

	deptID, _ := utils.GetDeptIDInt(c)

	dto, err := h.svc.Create(c.UserContext(), deptID, payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(dto)
}

func (h *PatientHandler) Update(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "clinic.update"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}

	var payload model.PatientDTO
	if err := c.BodyParser(&payload); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	payload.ID = id

	deptID, _ := utils.GetDeptIDInt(c)

	dto, err := h.svc.Update(c.UserContext(), deptID, payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *PatientHandler) Delete(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "clinic.delete"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
