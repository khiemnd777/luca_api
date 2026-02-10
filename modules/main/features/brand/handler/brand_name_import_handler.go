package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/brand/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type BrandNameImportHandler struct {
	svc  service.BrandNameImportService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewBrandNameImportHandler(svc service.BrandNameImportService, deps *module.ModuleDeps[config.ModuleConfig]) *BrandNameImportHandler {
	return &BrandNameImportHandler{svc: svc, deps: deps}
}

func (h *BrandNameImportHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/:dept_id<int>/brand/import-excel", h.ImportExcel)
}

func (h *BrandNameImportHandler) ImportExcel(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "product.create"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "file is required")
	}

	f, err := fileHeader.Open()
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "cannot open file")
	}
	defer f.Close()

	rows, err := service.ParseBrandNameExcel(f)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, fmt.Sprintf("invalid excel file: %v", err))
	}

	deptID, _ := utils.GetDeptIDInt(c)
	res, err := h.svc.ImportFromExcel(c.UserContext(), deptID, rows)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
