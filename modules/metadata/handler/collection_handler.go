package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/metadata/config"
	"github.com/khiemnd777/andy_api/modules/metadata/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
)

type CollectionHandler struct {
	svc  *service.CollectionService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewCollectionHandler(s *service.CollectionService, deps *module.ModuleDeps[config.ModuleConfig]) *CollectionHandler {
	return &CollectionHandler{svc: s, deps: deps}
}

// Mount dưới /metadata
func (h *CollectionHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/collections", h.List) // collections?query=&limit=&offset=&with_fields=true
	app.RouterPost(router, "/collections", h.Create)
	app.RouterGet(router, "/collections/:idOrSlug", h.GetOne)
	app.RouterPut(router, "/collections/:id", h.Update)
	app.RouterDelete(router, "/collections/:id", h.Delete)
}

func (h *CollectionHandler) List(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "privilege.metadata"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	in := service.ListCollectionsInput{
		Query:      c.Query("query"),
		Limit:      c.QueryInt("limit", 20),
		Offset:     c.QueryInt("offset", 0),
		WithFields: c.QueryBool("with_fields", false),
	}
	items, total, err := h.svc.List(c.UserContext(), in)
	if err != nil {
		logger.Error("collections.list failed", "err", err)
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": items, "total": total})
}

func (h *CollectionHandler) Create(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "privilege.metadata"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	var in service.CreateCollectionInput
	if err := c.BodyParser(&in); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	out, err := h.svc.Create(c.UserContext(), in)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *CollectionHandler) GetOne(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "privilege.metadata"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	withFields := c.QueryBool("withFields", false)
	idOrSlug := c.Params("idOrSlug")
	// nếu là số → ID
	if id, err := strconv.Atoi(idOrSlug); err == nil {
		out, err := h.svc.GetByID(c.UserContext(), id, withFields)
		if err != nil {
			return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
		}
		return c.JSON(out)
	}
	// slug
	out, err := h.svc.GetBySlug(c.UserContext(), idOrSlug, withFields)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *CollectionHandler) Update(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "privilege.metadata"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid id")
	}
	var in service.UpdateCollectionInput
	if err := c.BodyParser(&in); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	out, err := h.svc.Update(c.UserContext(), id, in)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *CollectionHandler) Delete(c *fiber.Ctx) error {
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
