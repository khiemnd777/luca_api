package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_statuses/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type CaseStatusesHandler struct {
	svc  service.CaseStatusesService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseStatusesHandler(svc service.CaseStatusesService, deps *module.ModuleDeps[config.ModuleConfig]) *CaseStatusesHandler {
	return &CaseStatusesHandler{svc: svc, deps: deps}
}

func (h *CaseStatusesHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/dashboard/case-statuses", h.CaseStatuses)
}

func (h *CaseStatusesHandler) CaseStatuses(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	deptID, _ := utils.GetDeptIDInt(c)

	res, err := h.svc.CaseStatuses(c.UserContext(), deptID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
