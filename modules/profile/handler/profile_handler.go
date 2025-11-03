package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/khiemnd777/andy_api/modules/profile/config"
	profileError "github.com/khiemnd777/andy_api/modules/profile/model"
	"github.com/khiemnd777/andy_api/modules/profile/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
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
	app.RouterGet(router, "/me", h.GetProfile)
	app.RouterPut(router, "/me/change-password", h.ChangePassword)
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

func (h *ProfileHandler) UpdateProfile(c *fiber.Ctx) error {
	if err := rbac.GuardRole(c, "user", h.deps.Ent.(*generated.Client)); err != nil {
		return err
	}

	type UpdateProfileRequest struct {
		Name       string  `json:"name"`
		Avatar     string  `json:"avatar"`
		Phone      *string `json:"phone"`
		Email      *string `json:"email"`
		BankQRCode *string `json:"bank_qr_code"`
	}

	body, err := app.ParseBody[UpdateProfileRequest](c)
	if err != nil {
		return err
	}

	userID, _ := utils.GetUserIDInt(c)
	updated, err := h.svc.UpdateProfile(c.UserContext(), userID, body.Name, body.Avatar, body.Phone, body.Email, body.BankQRCode)

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
	if err := rbac.GuardRole(c, "user", h.deps.Ent.(*generated.Client)); err != nil {
		return err
	}

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
	if err := rbac.GuardRole(c, "user", h.deps.Ent.(*generated.Client)); err != nil {
		return err
	}

	userID, _ := utils.GetUserIDInt(c)

	if err := h.svc.Delete(c.UserContext(), userID); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
