package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/metadata/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type FieldHandler struct{ svc *service.FieldService }

func NewFieldHandler(s *service.FieldService) *FieldHandler { return &FieldHandler{svc: s} }

// Mount dưới /metadata
func (h *FieldHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/fields", h.ListByCollection) // ?collection_id=
	app.RouterPost(router, "/fields", h.Create)
	app.RouterGet(router, "/fields/:id", h.Get)
	app.RouterPut(router, "/fields/:id", h.Update)
	app.RouterDelete(router, "/fields/:id", h.Delete)
}

func (h *FieldHandler) ListByCollection(c *fiber.Ctx) error {
	cid, err := strconv.Atoi(c.Query("collection_id", "0"))
	if err != nil || cid <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid collection_id")
	}
	out, err := h.svc.ListByCollection(c.Context(), cid)
	if err != nil {
		logger.Error("fields.list failed", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list fields")
	}
	return c.JSON(fiber.Map{"data": out})
}

func (h *FieldHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	out, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "field not found")
	}
	return c.JSON(out)
}

func (h *FieldHandler) Create(c *fiber.Ctx) error {
	var in service.FieldInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	out, err := h.svc.Create(c.Context(), in)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(out)
}

func (h *FieldHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var in service.FieldInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	out, err := h.svc.Update(c.Context(), id, in)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(out)
}

func (h *FieldHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "failed to delete")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
