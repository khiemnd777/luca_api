package handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/auth_facebook/config"
	"github.com/khiemnd777/andy_api/modules/auth_facebook/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/module"
)

type AuthFacebookHandler struct {
	service *service.AuthFacebookService
	deps    *module.ModuleDeps[config.ModuleConfig]
}

func NewAuthFacebookHandler(svc *service.AuthFacebookService, deps *module.ModuleDeps[config.ModuleConfig]) *AuthFacebookHandler {
	return &AuthFacebookHandler{
		service: svc,
		deps:    deps,
	}
}

func (h *AuthFacebookHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/", h.Login)
	app.RouterPost(router, "/limited", h.AuthLimited)
}

type FacebookLoginRequest struct {
	AccessToken string `json:"accessToken"`
}

func (h *AuthFacebookHandler) Login(c *fiber.Ctx) error {
	var req FacebookLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request body")
	}

	tokens, err := h.service.LoginWithFacebook(context.Background(), req.AccessToken)

	if err != nil {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, err, err.Error())
	}

	return c.JSON(tokens)
}

func (h *AuthFacebookHandler) AuthLimited(c *fiber.Ctx) error {
	type req struct {
		JWT string `json:"jwt"`
	}
	var r req
	if err := c.BodyParser(&r); err != nil || r.JWT == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing jwt")
	}

	tokens, err := h.service.FBLimitedVerify(c.UserContext(), r.JWT)

	if err != nil {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, err, err.Error())
	}

	return c.JSON(tokens)
}
