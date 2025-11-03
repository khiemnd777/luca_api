package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/auth_apple/config"
	"github.com/khiemnd777/andy_api/modules/auth_apple/model"
	"github.com/khiemnd777/andy_api/modules/auth_apple/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/module"
)

type AuthAppleHandler struct {
	service *service.AuthAppleService
	deps    *module.ModuleDeps[config.ModuleConfig]
}

func NewAuthAppleHandler(svc *service.AuthAppleService, deps *module.ModuleDeps[config.ModuleConfig]) *AuthAppleHandler {
	return &AuthAppleHandler{service: svc, deps: deps}
}

func (h *AuthAppleHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/", h.Login)
}

func (h *AuthAppleHandler) Login(c *fiber.Ctx) error {
	var req model.AppleUserInfo
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	tokenPair, err := h.service.LoginWithAppleUserInfo(c.UserContext(), &model.AppleUserInfo{
		Email:   req.Email,
		Name:    req.Name,
		Picture: req.Picture,
		Sub:     req.Sub,
	})
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, err, err.Error())
	}

	return c.JSON(tokenPair)
}
