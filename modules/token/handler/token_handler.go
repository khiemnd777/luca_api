package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/token/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
)

type TokenHandler struct {
	svc *service.TokenService
}

func NewTokenHandler(svc *service.TokenService) *TokenHandler {
	return &TokenHandler{svc: svc}
}

func (h *TokenHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/", h.Token)
	app.RouterPost(router, "/refresh", h.RefreshTokens)
	app.RouterPost(router, "/generate", h.GenerateTokens)
	app.RouterDelete(router, "/delete", h.DeleteRefreshToken)
}

func (h *TokenHandler) Token(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *TokenHandler) RefreshTokens(c *fiber.Ctx) error {
	type req struct {
		Token string `json:"refreshToken"`
	}
	var body req
	if err := c.BodyParser(&body); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request")
	}
	tokens, err := h.svc.RefreshToken(c.UserContext(), body.Token)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, err, err.Error())
	}

	return c.JSON(tokens)
}

func (h *TokenHandler) GenerateTokens(c *fiber.Ctx) error {
	type req struct {
		UserID int `json:"userID"`
	}

	var body req
	if err := c.BodyParser(&body); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request")
	}

	tokens, err := h.svc.GenerateTokens(c.UserContext(), body.UserID)

	if err != nil {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, err, err.Error())
	}

	return c.JSON(tokens)
}

func (h *TokenHandler) DeleteRefreshToken(c *fiber.Ctx) error {
	type req struct {
		RefreshToken string `json:"refreshToken"`
	}
	var body req
	if err := c.BodyParser(&body); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request")
	}
	if err := h.svc.DeleteRefreshToken(c.UserContext(), body.RefreshToken); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
