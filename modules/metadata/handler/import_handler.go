package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"

	"github.com/khiemnd777/andy_api/modules/metadata/config"
	"github.com/khiemnd777/andy_api/modules/metadata/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
)

type ImportHandler struct {
	engine *service.ImportEngine
	deps   *module.ModuleDeps[config.ModuleConfig]
}

func NewImportHandler(engine *service.ImportEngine, deps *module.ModuleDeps[config.ModuleConfig]) *ImportHandler {
	return &ImportHandler{engine: engine, deps: deps}
}

func (h *ImportHandler) RegisterRoutes(r fiber.Router) {
	// POST /metadata/import?scope=clinics&code=...
	app.RouterPost(r, "/import", h.Import)
}

func (h *ImportHandler) Import(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "privilege.metadata"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	scope := c.Query("scope")
	code := c.Query("code")
	if scope == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, fmt.Errorf("scope is required"), "scope is required")
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

	x, err := excelize.OpenReader(f)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid excel file")
	}
	defer x.Close()

	ctx := c.UserContext()

	// resolve profile + mappings
	profile, mappings, err := h.engine.Mapper.ResolveProfileAndMappings(c, ctx, scope, code)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	sheet := x.GetSheetName(0)
	if sheet == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, fmt.Errorf("empty sheet"), "empty sheet")
	}

	rows, err := x.Rows(sheet)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "cannot read rows")
	}
	defer rows.Close()

	rowIndex := 0
	created := 0
	updated := 0
	var errs []string

	for rows.Next() {
		rowIndex++
		cols, err := rows.Columns()
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: cannot read columns: %v", rowIndex, err))
			continue
		}

		if rowIndex == 1 {
			continue
		}

		valuesByCol := map[int]string{}
		for i, cell := range cols {
			valuesByCol[i+1] = cell
		}

		mapped, err := h.engine.Mapper.MapExcelRow(ctx, profile.ID, valuesByCol)
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: map error: %v", rowIndex, err))
			continue
		}

		isCreated, err := h.engine.UpsertMappedRow(ctx, scope, profile, mappings, mapped)
		if err != nil {
			errs = append(errs, fmt.Sprintf("row %d: upsert error: %v", rowIndex, err))
			continue
		}
		if isCreated {
			created++
		} else {
			updated++
		}
	}

	logger.Info("metadata.import.done", "scope", scope, "created", created, "updated", updated)

	return c.JSON(fiber.Map{
		"created": created,
		"updated": updated,
		"errors":  errs,
	})
}
