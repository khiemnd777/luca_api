package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/attribute/config"
	attribute "github.com/khiemnd777/andy_api/modules/attribute/model"
	"github.com/khiemnd777/andy_api/modules/attribute/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type AttributeOptionHandler struct {
	svc  *service.AttributeOptionService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewAttributeOptionHandler(svc *service.AttributeOptionService, deps *module.ModuleDeps[config.ModuleConfig]) *AttributeOptionHandler {
	return &AttributeOptionHandler{
		svc:  svc,
		deps: deps,
	}
}

func (h *AttributeOptionHandler) RegisterRoutes(router fiber.Router) {
	router.Route("/:id/options", func(r fiber.Router) {
		app.RouterGet(r, "/", h.ListOptions)
		app.RouterPost(r, "/", h.CreateOption)
		app.RouterPut(r, "/reorder", h.ReorderOptions)
		app.RouterPut(r, "/:option_id", h.UpdateOption)
		app.RouterDelete(r, "/:option_id", h.DeleteOption)
	})
}

func (h *AttributeOptionHandler) ListOptions(c *fiber.Ctx) error {
	attributeID, _ := strconv.Atoi(c.Params("id"))
	userID, _ := utils.GetUserIDInt(c)

	result, err := h.svc.ListOptions(c.UserContext(), userID, attributeID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(result)
}

func (h *AttributeOptionHandler) CreateOption(c *fiber.Ctx) error {
	attributeID, _ := strconv.Atoi(c.Params("id"))
	userID, _ := utils.GetUserIDInt(c)

	var req struct {
		OptionValue  string `json:"option_value"`
		DisplayOrder int    `json:"display_order"`
	}
	if err := c.BodyParser(&req); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	result, err := h.svc.CreateOption(c.UserContext(), userID, attributeID, req.OptionValue, req.DisplayOrder)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(result)
}

func (h *AttributeOptionHandler) UpdateOption(c *fiber.Ctx) error {
	attributeID, _ := strconv.Atoi(c.Params("id"))
	optionID, _ := strconv.Atoi(c.Params("option_id"))
	userID, _ := utils.GetUserIDInt(c)

	var req struct {
		OptionValue  string `json:"option_value"`
		DisplayOrder int    `json:"display_order"`
	}
	if err := c.BodyParser(&req); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	result, err := h.svc.UpdateOption(c.UserContext(), userID, attributeID, optionID, req.OptionValue, req.DisplayOrder)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(result)
}

func (h *AttributeOptionHandler) DeleteOption(c *fiber.Ctx) error {
	attributeID, _ := strconv.Atoi(c.Params("id"))
	optionID, _ := strconv.Atoi(c.Params("option_id"))
	userID, _ := utils.GetUserIDInt(c)

	if err := h.svc.DeleteOption(c.UserContext(), userID, attributeID, optionID); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AttributeOptionHandler) ReorderOptions(c *fiber.Ctx) error {
	attributeID, _ := strconv.Atoi(c.Params("id"))
	userID, _ := utils.GetUserIDInt(c)

	var req struct {
		Orders []attribute.OptionOrder `json:"orders"`
	}

	if err := c.BodyParser(&req); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	err := h.svc.BatchUpdateDisplayOrder(c.UserContext(), userID, attributeID, req.Orders)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.SendStatus(fiber.StatusOK)
}
