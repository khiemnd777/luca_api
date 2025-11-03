package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/user/config"
	"github.com/khiemnd777/andy_api/modules/user/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type UserHandler struct {
	svc  *service.UserService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewUserHandler(svc *service.UserService, deps *module.ModuleDeps[config.ModuleConfig]) *UserHandler {
	return &UserHandler{svc: svc, deps: deps}
}

func (h *UserHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/", h.Create)
	app.RouterGet(router, "/admin", h.GetAdminUserID)
	app.RouterPost(router, "/batch-get", h.BatchGet)
	app.RouterGet(router, "/qr/:ref_code", h.GetUserByRefCode)
	app.RouterGet(router, "/:id/qr", h.GetQRCodeByUserID)
	app.RouterGet(router, "/:id", h.GetByID)
	app.RouterPut(router, "/:id", h.Update)
	app.RouterDelete(router, "/:id", h.Delete)
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	type req struct {
		Email    string  `json:"email"`
		Password string  `json:"password"`
		Name     string  `json:"name"`
		Phone    *string `json:"phone"`
	}
	var body req
	if err := c.BodyParser(&body); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid body")
	}

	user, err := h.svc.Create(c.UserContext(), body.Email, body.Password, body.Name, body.Phone)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(user)
}

func (h *UserHandler) GetByID(c *fiber.Ctx) error {
	id, _ := utils.GetParamAsInt(c, "id")
	user, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, "User not found")
	}
	return c.JSON(user)
}

func (h *UserHandler) GetAdminUserID(c *fiber.Ctx) error {
	if err := rbac.GuardRole(c, "admin", h.deps.Ent.(*generated.Client)); err != nil {
		return err
	}
	userID, err := h.svc.GetAdminUserID(c.UserContext())
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, "not found")
	}
	return c.JSON(userID)
}

func (h *UserHandler) GetQRCodeByUserID(c *fiber.Ctx) error {
	userID, _ := utils.GetParamAsInt(c, "id")
	qrCode, err := h.svc.GetQRCodeByUserID(c.UserContext(), userID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "Failed to retrieve bank QR code")
	}
	if qrCode == nil {
		return c.SendStatus(fiber.StatusNoContent)
	}
	return c.JSON(fiber.Map{"data": qrCode})
}

func (h *UserHandler) GetUserByRefCode(c *fiber.Ctx) error {
	refCode := utils.GetParamAsString(c, "ref_code")

	user, err := h.svc.GetUserByRefCode(c.UserContext(), refCode)

	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "Failed to retrieve QR code")
	}
	if user == nil {
		return c.SendStatus(fiber.StatusNoContent)
	}
	return c.JSON(user)
}

func (h *UserHandler) BatchGet(c *fiber.Ctx) error {
	var req struct {
		IDs []int `json:"ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request body")
	}

	if len(req.IDs) == 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "IDs cannot be empty")
	}

	users, err := h.svc.BatchGetByIDs(c.UserContext(), req.IDs)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "Failed to fetch users")
	}

	return c.JSON(users)
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	if err := rbac.GuardRole(c, "user", h.deps.Ent.(*generated.Client)); err != nil {
		return err
	}

	id, _ := utils.GetParamAsInt(c, "id")
	type req struct {
		Name       string  `json:"name"`
		Phone      *string `json:"phone"`
		Avatar     *string `json:"avatar"`
		BankQRCode *string `json:"bank_qr_code"`
	}
	var body req
	if err := c.BodyParser(&body); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid body")
	}
	user, err := h.svc.Update(c.UserContext(), id, body.Name, body.Phone, body.Avatar, body.BankQRCode)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(user)
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	if err := rbac.GuardRole(c, "user", h.deps.Ent.(*generated.Client)); err != nil {
		return err
	}

	id, _ := utils.GetParamAsInt(c, "id")
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
