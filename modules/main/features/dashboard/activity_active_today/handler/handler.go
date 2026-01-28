package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/activity_active_today/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type ActiveTodayHandler struct {
	svc  service.ActiveTodayService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewActiveTodayHandler(svc service.ActiveTodayService, deps *module.ModuleDeps[config.ModuleConfig]) *ActiveTodayHandler {
	return &ActiveTodayHandler{svc: svc, deps: deps}
}

func (h *ActiveTodayHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/dashboard/active-today", h.ActiveToday)
}

func (h *ActiveTodayHandler) ActiveToday(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	deptID, _ := utils.GetDeptIDInt(c)

	res, err := h.svc.ActiveToday(
		c.UserContext(),
		deptID,
	)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}
