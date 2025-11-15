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

type FieldHandler struct {
	svc  *service.FieldService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewFieldHandler(s *service.FieldService, deps *module.ModuleDeps[config.ModuleConfig]) *FieldHandler {
	return &FieldHandler{svc: s, deps: deps}
}

// Mount dưới /metadata
func (h *FieldHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/fields", h.ListByCollection) // ?collection_id=
	app.RouterPost(router, "/fields", h.Create)
	app.RouterGet(router, "/fields/:id", h.Get)
	app.RouterPut(router, "/fields/:id", h.Update)
	app.RouterDelete(router, "/fields/:id", h.Delete)
}

func (h *FieldHandler) ListByCollection(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "privilege.metadata"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	cid, err := strconv.Atoi(c.Query("collection_id", "0"))
	if err != nil || cid <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid collection_id")
	}
	out, err := h.svc.ListByCollection(c.UserContext(), cid)
	if err != nil {
		logger.Error("fields.list failed", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list fields")
	}
	return c.JSON(fiber.Map{"data": out})
}

func (h *FieldHandler) Get(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "privilege.metadata"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	out, err := h.svc.Get(c.UserContext(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "field not found")
	}
	return c.JSON(out)
}

func (h *FieldHandler) Create(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "privilege.metadata"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	var in model.FieldInput
	if err := c.BodyParser(&in); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	out, err := h.svc.Create(c.UserContext(), in)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *FieldHandler) Update(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "privilege.metadata"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}
	var in model.FieldInput
	if err := c.BodyParser(&in); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	out, err := h.svc.Update(c.UserContext(), id, in)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *FieldHandler) Delete(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "privilege.metadata"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
