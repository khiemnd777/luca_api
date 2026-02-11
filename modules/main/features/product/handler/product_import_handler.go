package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/product/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type ProductImportHandler struct {
	svc  service.ProductImportService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewProductImportHandler(svc service.ProductImportService, deps *module.ModuleDeps[config.ModuleConfig]) *ProductImportHandler {
	return &ProductImportHandler{svc: svc, deps: deps}
}

func (h *ProductImportHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/:dept_id<int>/product/import-excel", h.ImportExcel)
}

func (h *ProductImportHandler) ImportExcel(c *fiber.Ctx) error {
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

	rows, err := service.ParseProductExcel(f)
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
