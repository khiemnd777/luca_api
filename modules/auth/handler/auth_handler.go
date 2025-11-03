package handler

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	authErrors "github.com/khiemnd777/andy_api/modules/auth/model/error"
	"github.com/khiemnd777/andy_api/modules/auth/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/login", h.Login)
	app.RouterPost(router, "/register", h.Register)
	app.RouterPost(router, "/refresh-token", h.Refresh)
	app.RouterPost(router, "/logout", h.Logout)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	type req struct {
		PhoneOrEmail string `json:"phone_or_email"`
		Password     string `json:"password"`
	}
	var body req
	if err := c.BodyParser(&body); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request")
	}

	logger.Debug(fmt.Sprintf("Attempting login for %s", body.PhoneOrEmail))

	tokens, err := h.svc.Login(c.UserContext(), body.PhoneOrEmail, body.Password)
	if err != nil {
		switch {
		case errors.Is(err, authErrors.ErrInvalidCredentials):
			return client_error.ResponseServiceMessage(c, client_error.ServiceMessageCode, err.Error())
		default:
			return client_error.ResponseError(c, fiber.StatusUnauthorized, err, err.Error())
		}
	}

	return c.JSON(tokens)
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	type RegisterRequest struct {
		PhoneOrEmail string `json:"phone_or_email"`
		Password     string `json:"password"`
		Name         string `json:"name"`
	}
	var body RegisterRequest
	if err := c.BodyParser(&body); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request")
	}

	_, err := h.svc.CreateNewUser(c.UserContext(), body.PhoneOrEmail, body.Password, body.Name)
	if err != nil {
		switch {
		case errors.Is(err, authErrors.ErrPhoneOrEmailExists):
			return client_error.ResponseServiceMessage(c, client_error.ServiceMessageCode, err.Error())
		default:
			return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
		}
	}

	// Auto login after register
	tokens, err := h.svc.Login(c.UserContext(), body.PhoneOrEmail, body.Password)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, err, err.Error())
	}

	return c.JSON(tokens)
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
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

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	type req struct {
		Token string `json:"refreshToken"`
	}
	var body req
	if err := c.BodyParser(&body); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request")
	}
	if err := h.svc.Logout(c.UserContext(), body.Token); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
