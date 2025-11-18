package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/metadata/config"
	"github.com/khiemnd777/andy_api/modules/metadata/model"
	"github.com/khiemnd777/andy_api/modules/metadata/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
)

type ImportFieldMappingHandler struct {
	svc  *service.ImportFieldMappingService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewImportFieldMappingHandler(
	svc *service.ImportFieldMappingService,
	deps *module.ModuleDeps[config.ModuleConfig],
) *ImportFieldMappingHandler {
	return &ImportFieldMappingHandler{svc: svc, deps: deps}
}

func (h *ImportFieldMappingHandler) RegisterRoutes(r fiber.Router) {
	app.RouterGet(r, "/import-mappings", h.ListByProfile)
	app.RouterPost(r, "/import-mappings", h.Create)
	app.RouterGet(r, "/import-mappings/:id", h.Get)
	app.RouterPut(r, "/import-mappings/:id", h.Update)
	app.RouterDelete(r, "/import-mappings/:id", h.Delete)
}

func (h *ImportFieldMappingHandler) guard(c *fiber.Ctx) error {
	return rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "privilege.metadata")
}

// GET /import-mappings?profile_id=1
func (h *ImportFieldMappingHandler) ListByProfile(c *fiber.Ctx) error {
	if err := h.guard(c); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	pid, err := strconv.Atoi(c.Query("profile_id", "0"))
	if err != nil || pid <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "profile_id is required")
	}

	out, err := h.svc.ListByProfileID(c.UserContext(), pid)
	if err != nil {
		logger.Error("import_mappings.list failed", "err", err)
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "failed to list import mappings")
	}
	return c.JSON(fiber.Map{"data": out})
}

func (h *ImportFieldMappingHandler) Get(c *fiber.Ctx) error {
	if err := h.guard(c); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}

	out, err := h.svc.Get(c.UserContext(), id)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, "mapping not found")
	}
	return c.JSON(out)
}

func (h *ImportFieldMappingHandler) Create(c *fiber.Ctx) error {
	if err := h.guard(c); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	var in model.ImportFieldMappingInput
	if err := c.BodyParser(&in); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}

	out, err := h.svc.Create(c.UserContext(), in)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *ImportFieldMappingHandler) Update(c *fiber.Ctx) error {
	if err := h.guard(c); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}

	var in model.ImportFieldMappingInput
	if err := c.BodyParser(&in); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}

	out, err := h.svc.Update(c.UserContext(), id, in)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *ImportFieldMappingHandler) Delete(c *fiber.Ctx) error {
	if err := h.guard(c); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}

	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
