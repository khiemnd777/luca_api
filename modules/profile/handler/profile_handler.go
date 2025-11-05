package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/profile/config"
	profileError "github.com/khiemnd777/andy_api/modules/profile/model"
	"github.com/khiemnd777/andy_api/modules/profile/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type ProfileHandler struct {
	svc  *service.ProfileService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewProfileHandler(svc *service.ProfileService, deps *module.ModuleDeps[config.ModuleConfig]) *ProfileHandler {
	return &ProfileHandler{
		svc:  svc,
		deps: deps,
	}
}

func (h *ProfileHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPut(router, "/me/change-password", h.ChangePassword)
	app.RouterPost(router, "/me/exists-email", h.ExistsEmail)
	app.RouterPost(router, "/me/exists-phone", h.ExistsPhone)

	app.RouterGet(router, "/me", h.GetProfile)
	app.RouterPut(router, "/me", h.UpdateProfile)
	app.RouterDelete(router, "/me", h.Delete)
}

func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	userID, _ := utils.GetUserIDInt(c)
	profile, err := h.svc.GetProfile(c.UserContext(), userID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusNotFound, err, "User not found")
	}
	return c.JSON(profile)
}

func (h *ProfileHandler) ExistsEmail(c *fiber.Ctx) error {
	type ExistsEmailRequest struct {
		Email string `json:"email"`
	}

	body, err := app.ParseBody[ExistsEmailRequest](c)
	if err != nil {
		return err
	}
	userID, _ := utils.GetUserIDInt(c)
	existed, err := h.svc.CheckEmailExists(c.UserContext(), userID, body.Email)

	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(existed)
}

func (h *ProfileHandler) ExistsPhone(c *fiber.Ctx) error {
	type ExistsPhoneRequest struct {
		Phone string `json:"phone"`
	}

	body, err := app.ParseBody[ExistsPhoneRequest](c)
	if err != nil {
		return err
	}
	userID, _ := utils.GetUserIDInt(c)
	existed, err := h.svc.CheckPhoneExists(c.UserContext(), userID, body.Phone)

	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.JSON(existed)
}

func (h *ProfileHandler) UpdateProfile(c *fiber.Ctx) error {
	type UpdateProfileRequest struct {
		ID     int     `json:"id"`
		Name   string  `json:"name"`
		Avatar string  `json:"avatar"`
		Phone  *string `json:"phone"`
		Email  *string `json:"email"`
	}

	body, err := app.ParseBody[UpdateProfileRequest](c)
	if err != nil {
		return err
	}

	userID, _ := utils.GetUserIDInt(c)
	updated, err := h.svc.UpdateProfile(c.UserContext(), userID, body.Name, body.Avatar, body.Phone, body.Email)

	if err != nil {
		switch {
		case errors.Is(err, profileError.ErrEmailExists):
		case errors.Is(err, profileError.ErrPhoneExists):
			return client_error.ResponseServiceMessage(c, client_error.ServiceMessageCode, err.Error())
		default:
			return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
		}
	}

	return c.JSON(updated)
}

func (h *ProfileHandler) ChangePassword(c *fiber.Ctx) error {
	type ChangePasswordRequest struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	var body ChangePasswordRequest
	if err := c.BodyParser(&body); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request body")
	}

	userID, _ := utils.GetUserIDInt(c)

	if err := h.svc.ChangePassword(c.UserContext(), userID, body.CurrentPassword, body.NewPassword); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.SendStatus(fiber.StatusAccepted)
}

func (h *ProfileHandler) Delete(c *fiber.Ctx) error {
	userID, _ := utils.GetUserIDInt(c)

	if err := h.svc.Delete(c.UserContext(), userID); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
