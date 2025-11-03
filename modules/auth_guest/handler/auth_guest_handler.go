package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/auth_guest/config"
	"github.com/khiemnd777/andy_api/modules/auth_guest/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/module"
)

type AuthGuestHandler struct {
	service *service.AuthGuestService
	deps    *module.ModuleDeps[config.ModuleConfig]
}

func NewAuthGuestHandler(svc *service.AuthGuestService, deps *module.ModuleDeps[config.ModuleConfig]) *AuthGuestHandler {
	return &AuthGuestHandler{service: svc, deps: deps}
}

func (h *AuthGuestHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/", h.Login)
	app.RouterDelete(router, "/", h.Delete)
}

func (h *AuthGuestHandler) Login(c *fiber.Ctx) error {
	tokenPair, err := h.service.LoginWithGuest(c.UserContext())
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, err, err.Error())
	}

	return c.JSON(tokenPair)
}

func (h *AuthGuestHandler) Delete(c *fiber.Ctx) error {
	type GuestDeleteRequest struct {
		UserID int `json:"user_id"`
	}

	body, err := app.ParseBody[GuestDeleteRequest](c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	err = h.service.DeleteGuest(c.UserContext(), body.UserID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}
