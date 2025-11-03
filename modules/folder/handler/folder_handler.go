// scripts/create_module/templates/handler_http.go.tmpl
package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/folder/config"
	"github.com/khiemnd777/andy_api/modules/folder/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type FolderHandler struct {
	svc  *service.FolderService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewFolderHandler(svc *service.FolderService, deps *module.ModuleDeps[config.ModuleConfig]) *FolderHandler {
	return &FolderHandler{
		svc:  svc,
		deps: deps,
	}
}

func (h *FolderHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/", h.Create)
	app.RouterGet(router, "/", h.List)
	app.RouterGet(router, "/:id", h.Get)
	app.RouterPut(router, "/:id", h.Update)
	app.RouterDelete(router, "/:id", h.Delete)
}

func (h *FolderHandler) Create(c *fiber.Ctx) error {
	var input generated.Folder
	if err := c.BodyParser(&input); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid input")
	}
	userId, _ := utils.GetUserIDInt(c)
	folder, err := h.svc.Create(c.UserContext(), userId, &input)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(folder)
}

func (h *FolderHandler) Get(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	userId, _ := utils.GetUserIDInt(c)
	folder, err := h.svc.Get(c.UserContext(), id, userId)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, "Folder not found")
	}
	return c.JSON(folder)
}

func (h *FolderHandler) List(c *fiber.Ctx) error {
	userId, _ := strconv.Atoi(c.Query("user_id"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	items, hasMore, err := h.svc.ListPaginated(c.UserContext(), userId, page, limit)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(fiber.Map{
		"items":   items,
		"hasMore": hasMore,
	})
}

func (h *FolderHandler) Update(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var input generated.Folder
	if err := c.BodyParser(&input); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid input")
	}
	userId, _ := utils.GetUserIDInt(c)
	updated, err := h.svc.Update(c.UserContext(), id, userId, &input)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(updated)
}

func (h *FolderHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	userId, _ := utils.GetUserIDInt(c)
	if err := h.svc.Delete(c.UserContext(), id, userId); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
