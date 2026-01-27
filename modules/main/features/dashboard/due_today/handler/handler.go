package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/due_today/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type DueTodayHandler struct {
	svc  service.DueTodayService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewDueTodayHandler(svc service.DueTodayService, deps *module.ModuleDeps[config.ModuleConfig]) *DueTodayHandler {
	return &DueTodayHandler{svc: svc, deps: deps}
}

func (h *DueTodayHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/dashboard/due-today", h.DueToday)
}

func (h *DueTodayHandler) DueToday(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	deptID, _ := utils.GetDeptIDInt(c)

	res, err := h.svc.DueToday(
		c.UserContext(),
		deptID,
	)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}
