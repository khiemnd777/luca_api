package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/metadata/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type CollectionHandler struct{ svc *service.CollectionService }

func NewCollectionHandler(s *service.CollectionService) *CollectionHandler {
	return &CollectionHandler{svc: s}
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
	in := service.ListCollectionsInput{
		Query:      c.Query("query"),
		Limit:      c.QueryInt("limit", 20),
		Offset:     c.QueryInt("offset", 0),
		WithFields: c.QueryBool("with_fields", false),
	}
	items, total, err := h.svc.List(c.UserContext(), in)
	if err != nil {
		logger.Error("collections.list failed", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list collections")
	}
	return c.JSON(fiber.Map{"data": items, "total": total})
}

func (h *CollectionHandler) Create(c *fiber.Ctx) error {
	var in service.CreateCollectionInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	out, err := h.svc.Create(c.UserContext(), in)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(out)
}

func (h *CollectionHandler) GetOne(c *fiber.Ctx) error {
	withFields := c.QueryBool("withFields", false)
	idOrSlug := c.Params("idOrSlug")
	// nếu là số → ID
	if id, err := strconv.Atoi(idOrSlug); err == nil {
		out, err := h.svc.GetByID(c.UserContext(), id, withFields)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "collection not found")
		}
		return c.JSON(out)
	}
	// slug
	out, err := h.svc.GetBySlug(c.UserContext(), idOrSlug, withFields)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "collection not found")
	}
	return c.JSON(out)
}

func (h *CollectionHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var in service.UpdateCollectionInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	out, err := h.svc.Update(c.UserContext(), id, in)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(out)
}

func (h *CollectionHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "failed to delete")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
