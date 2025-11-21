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
	app.RouterGet(router, "/collections", h.List) // collections?query=&limit=&offset=&with_fields=true&table=false&form=false
	app.RouterPost(router, "/collections", h.Create)
	app.RouterGet(router, "/collections/:idOrSlug", h.GetOne)
	app.RouterPost(router, "/collections/available/:idOrSlug", h.GetAvailableOne)
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
		Table:      c.QueryBool("table", false),
		Form:       c.QueryBool("form", false),
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
	table := c.QueryBool("table", false)
	form := c.QueryBool("form", false)
	idOrSlug := c.Params("idOrSlug")
	// ID
	if id, err := strconv.Atoi(idOrSlug); err == nil {
		out, err := h.svc.GetByID(c.UserContext(), id, withFields, table, form)
		if err != nil {
			return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
		}
		return c.JSON(out)
	}
	// slug
	out, err := h.svc.GetBySlug(c.UserContext(), idOrSlug, withFields, table, form)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(out)
}

func (h *CollectionHandler) GetAvailableOne(c *fiber.Ctx) error {
	withFields := c.QueryBool("withFields", false)
	table := c.QueryBool("table", false)
	form := c.QueryBool("form", false)
	idOrSlug := c.Params("idOrSlug")

	var entityData map[string]any
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&entityData); err != nil {
			return client_error.ResponseError(
				c,
				fiber.StatusBadRequest,
				err,
				"invalid body",
			)
		}
	}

	var entityPtr *map[string]any
	if entityData != nil {
		entityPtr = &entityData
	}

	// ID
	if id, err := strconv.Atoi(idOrSlug); err == nil {
		out, err := h.svc.GetAvailableByID(c.UserContext(), id, withFields, table, form, entityPtr)
		if err != nil {
			return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
		}
		return c.JSON(out)
	}
	// slug
	out, err := h.svc.GetByAvailableSlug(c.UserContext(), idOrSlug, withFields, table, form, entityPtr)
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
