package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/dashboard/case_daily_sales_stats/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type CaseDailySalesStatsHandler struct {
	svc  service.CaseDailySalesStatsService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseDailySalesStatsHandler(svc service.CaseDailySalesStatsService, deps *module.ModuleDeps[config.ModuleConfig]) *CaseDailySalesStatsHandler {
	return &CaseDailySalesStatsHandler{svc: svc, deps: deps}
}

func (h *CaseDailySalesStatsHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/dashboard/case-daily-sales-stats/summary", h.Summary)
	app.RouterGet(router, "/:dept_id<int>/dashboard/case-daily-sales-stats/daily", h.Daily)
}

func (h *CaseDailySalesStatsHandler) Summary(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	deptID, err := resolveDepartmentID(c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	fromDateRaw := utils.GetQueryAsString(c, "from_date")
	if fromDateRaw == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid from_date")
	}
	fromDate, err := utils.ParseDate(fromDateRaw)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid from_date")
	}

	toDateRaw := utils.GetQueryAsString(c, "to_date")
	if toDateRaw == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid to_date")
	}
	toDate, err := utils.ParseDate(toDateRaw)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid to_date")
	}

	previousFromRaw := utils.GetQueryAsString(c, "previous_from_date")
	if previousFromRaw == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid previous_from_date")
	}
	previousFrom, err := utils.ParseDate(previousFromRaw)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid previous_from_date")
	}

	previousToRaw := utils.GetQueryAsString(c, "previous_to_date")
	if previousToRaw == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid previous_to_date")
	}
	previousTo, err := utils.ParseDate(previousToRaw)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid previous_to_date")
	}

	res, err := h.svc.Summary(
		c.UserContext(),
		deptID,
		fromDate,
		toDate,
		previousFrom,
		previousTo,
	)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *CaseDailySalesStatsHandler) Daily(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "order.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	deptID, err := resolveDepartmentID(c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, err.Error())
	}

	fromDateRaw := utils.GetQueryAsString(c, "from_date")
	if fromDateRaw == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid from_date")
	}
	fromDate, err := utils.ParseDate(fromDateRaw)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid from_date")
	}

	toDateRaw := utils.GetQueryAsString(c, "to_date")
	if toDateRaw == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid to_date")
	}
	toDate, err := utils.ParseDate(toDateRaw)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid to_date")
	}

	res, err := h.svc.Daily(c.UserContext(), deptID, fromDate, toDate)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func resolveDepartmentID(c *fiber.Ctx) (int, error) {
	departmentID, err := utils.GetQueryAsNillableInt(c, "department_id")
	if err != nil {
		return 0, err
	}
	if departmentID != nil && *departmentID <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid department_id")
	}
	if departmentID == nil {
		paramDeptID, err := utils.GetParamAsNillableInt(c, "dept_id")
		if err != nil {
			return 0, err
		}
		if paramDeptID != nil {
			departmentID = paramDeptID
		}
	}
	if departmentID == nil {
		if deptID, ok := utils.GetDeptIDInt(c); ok && deptID > 0 {
			departmentID = &deptID
		}
	}
	if departmentID == nil || *departmentID <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid department_id")
	}
	return *departmentID, nil
}
