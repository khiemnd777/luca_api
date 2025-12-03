package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/staff/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type StaffHandler struct {
	svc  service.StaffService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewStaffHandler(svc service.StaffService, deps *module.ModuleDeps[config.ModuleConfig]) *StaffHandler {
	return &StaffHandler{svc: svc, deps: deps}
}

func (h *StaffHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/staff/list", h.List)
	app.RouterGet(router, "/:dept_id<int>/staff/search", h.Search)
	app.RouterGet(router, "/:dept_id<int>/staff/role/:role_name/search", h.SearchWithRoleName)
	app.RouterGet(router, "/:dept_id<int>/section/:section_id<int>/staffs", h.ListBySectionID)
	app.RouterGet(router, "/:dept_id<int>/role/:role_name/staffs", h.ListByRoleName)
	app.RouterGet(router, "/:dept_id<int>/staff/:id<int>", h.GetByID)
	app.RouterPost(router, "/:dept_id<int>/staff", h.Create)
	app.RouterPost(router, "/:dept_id<int>/staff/change-password", h.ChangePassword)
	app.RouterPut(router, "/:dept_id<int>/staff/:id<int>", h.Update)
	app.RouterPost(router, "/:dept_id<int>/staff/:id<int>/exists-phone", h.ExistsPhone)
	app.RouterPost(router, "/:dept_id<int>/staff/:id<int>/exists-email", h.ExistsEmail)
	app.RouterDelete(router, "/:dept_id<int>/staff/:id<int>", h.Delete)
}

func (h *StaffHandler) List(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "staff.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := table.ParseTableQuery(c, 20)
	res, err := h.svc.List(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *StaffHandler) ListBySectionID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "staff.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := table.ParseTableQuery(c, 20)
	sectionID, _ := utils.GetParamAsInt(c, "section_id")
	res, err := h.svc.ListBySectionID(c.UserContext(), sectionID, q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *StaffHandler) ListByRoleName(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "staff.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := table.ParseTableQuery(c, 20)
	roleName := utils.GetParamAsString(c, "role_name")
	res, err := h.svc.ListByRoleName(c.UserContext(), roleName, q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *StaffHandler) Search(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "staff.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := dbutils.ParseSearchQuery(c, 20)
	res, err := h.svc.Search(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *StaffHandler) SearchWithRoleName(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "staff.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	q := dbutils.ParseSearchQuery(c, 20)
	roleName := utils.GetParamAsString(c, "role_name")
	res, err := h.svc.SearchWithRoleName(c.UserContext(), roleName, q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *StaffHandler) GetByID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "staff.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}

	dto, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *StaffHandler) ExistsEmail(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "staff.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id < -1 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}

	type ExistsEmailRequest struct {
		Email string `json:"email"`
	}

	body, err := app.ParseBody[ExistsEmailRequest](c)
	if err != nil {
		return err
	}

	existed, err := h.svc.CheckEmailExists(c.UserContext(), id, body.Email)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.
		Status(fiber.StatusOK).
		JSON(existed)
}

func (h *StaffHandler) ExistsPhone(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "staff.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id < -1 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}

	type ExistsPhoneRequest struct {
		Phone string `json:"phone"`
	}

	body, err := app.ParseBody[ExistsPhoneRequest](c)
	if err != nil {
		return err
	}

	existed, err := h.svc.CheckPhoneExists(c.UserContext(), id, body.Phone)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.
		Status(fiber.StatusOK).
		JSON(existed)
}

func (h *StaffHandler) Create(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "staff.create"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	var payload model.StaffDTO
	if err := c.BodyParser(&payload); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	if payload.Name == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "name is required")
	}

	deptID, _ := utils.GetDeptIDInt(c)

	dto, err := h.svc.Create(c.UserContext(), deptID, payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(dto)
}

func (h *StaffHandler) Update(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "staff.update"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}

	var payload model.StaffDTO
	if err := c.BodyParser(&payload); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "invalid body")
	}
	payload.ID = id

	deptID, _ := utils.GetDeptIDInt(c)

	dto, err := h.svc.Update(c.UserContext(), deptID, payload)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(dto)
}

func (h *StaffHandler) ChangePassword(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "staff.update"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	type ChangePasswordRequest struct {
		NewPassword string `json:"new_password"`
	}

	var body ChangePasswordRequest
	if err := c.BodyParser(&body); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "Invalid request body")
	}

	userID, _ := utils.GetUserIDInt(c)

	if err := h.svc.ChangePassword(c.UserContext(), userID, body.NewPassword); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.SendStatus(fiber.StatusAccepted)
}

func (h *StaffHandler) Delete(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "staff.delete"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid id")
	}
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
