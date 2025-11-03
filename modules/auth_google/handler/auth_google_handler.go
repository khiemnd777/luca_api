package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/auth_google/config"
	"github.com/khiemnd777/andy_api/modules/auth_google/model"
	"github.com/khiemnd777/andy_api/modules/auth_google/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/module"
)

type AuthGoogleHandler struct {
	service *service.AuthGoogleService
	deps    *module.ModuleDeps[config.ModuleConfig]
}

func NewAuthGoogleHandler(svc *service.AuthGoogleService, deps *module.ModuleDeps[config.ModuleConfig]) *AuthGoogleHandler {
	return &AuthGoogleHandler{service: svc, deps: deps}
}

func (h *AuthGoogleHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/", h.Login)
}

type GoogleLoginRequest struct {
	IDToken string `json:"idToken"`
}

func (h *AuthGoogleHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
		Sub     string `json:"sub"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	tokenPair, err := h.service.LoginWithGoogleUserInfo(c.UserContext(), &model.GoogleUserInfo{
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
