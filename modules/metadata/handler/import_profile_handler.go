package handler

import (
	"strconv"
	"strings"

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

type ImportFieldProfileHandler struct {
	svc  *service.ImportFieldProfileService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewImportFieldProfileHandler(
	svc *service.ImportFieldProfileService,
	deps *module.ModuleDeps[config.ModuleConfig],
) *ImportFieldProfileHandler {
	return &ImportFieldProfileHandler{svc: svc, deps: deps}
}

func (h *ImportFieldProfileHandler) RegisterRoutes(r fiber.Router) {
	app.RouterGet(r, "/import-profiles", h.List)
	app.RouterPost(r, "/import-profiles", h.Create)
	app.RouterGet(r, "/import-profiles/:id", h.Get)
	app.RouterPut(r, "/import-profiles/:id", h.Update)
	app.RouterDelete(r, "/import-profiles/:id", h.Delete)
}

func (h *ImportFieldProfileHandler) guard(c *fiber.Ctx) error {
	return rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "privilege.metadata")
}

func (h *ImportFieldProfileHandler) List(c *fiber.Ctx) error {
	if err := h.guard(c); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	scope := strings.TrimSpace(c.Query("scope"))

	out, err := h.svc.List(c.UserContext(), scope)
	if err != nil {
		logger.Error("import_profiles.list failed", "err", err)
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "failed to list import profiles")
	}
	return c.JSON(fiber.Map{"data": out})
}

func (h *ImportFieldProfileHandler) Get(c *fiber.Ctx) error {
	if err := h.guard(c); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}

	out, err := h.svc.Get(c.UserContext(), id)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, "profile not found")
	}
	return c.JSON(out)
}

func (h *ImportFieldProfileHandler) Create(c *fiber.Ctx) error {
	if err := h.guard(c); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	var in model.ImportFieldProfileInput
	if err := c.BodyParser(&in); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}

	out, err := h.svc.Create(c.UserContext(), in)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *ImportFieldProfileHandler) Update(c *fiber.Ctx) error {
	if err := h.guard(c); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}

	var in model.ImportFieldProfileInput
	if err := c.BodyParser(&in); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}

	out, err := h.svc.Update(c.UserContext(), id, in)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *ImportFieldProfileHandler) Delete(c *fiber.Ctx) error {
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
