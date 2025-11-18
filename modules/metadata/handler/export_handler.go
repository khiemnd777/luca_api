package handler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"

	"github.com/khiemnd777/andy_api/modules/metadata/config"
	"github.com/khiemnd777/andy_api/modules/metadata/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/module"
)

type ExportHandler struct {
	engine *service.ImportEngine
	deps   *module.ModuleDeps[config.ModuleConfig]
}

func NewExportHandler(engine *service.ImportEngine, deps *module.ModuleDeps[config.ModuleConfig]) *ExportHandler {
	return &ExportHandler{engine: engine, deps: deps}
}

func (h *ExportHandler) RegisterRoutes(r fiber.Router) {
	// POST /metadata/export?scope=clinics&code=...
	app.RouterPost(r, "/export", h.Export)
}

func (h *ExportHandler) Export(c *fiber.Ctx) error {
	scope := strings.TrimSpace(c.Query("scope"))
	code := strings.TrimSpace(c.Query("code"))
	if scope == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, fmt.Errorf("scope is required"), "scope is required")
	}

	ctx := c.UserContext()

	// 1) Resolve profile + mappings (dựa trên scope + code hoặc default)
	profile, mappings, err := h.engine.Mapper.ResolveProfileAndMappings(c, ctx, scope, code)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	// 2) Tập hợp danh sách core columns cần SELECT
	coreCols := map[string]struct{}{}
	for _, m := range mappings {
		if strings.ToLower(m.InternalKind) == "core" && strings.TrimSpace(m.InternalPath) != "" {
			coreCols[m.InternalPath] = struct{}{}
		}
	}

	tableName, err := h.engine.SafeIdent(scope)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid scope")
	}

	// SELECT id, custom_fields, core columns...
	colList := []string{`id`, `custom_fields`}
	for col := range coreCols {
		ident, err := h.engine.SafeIdent(col)
		if err != nil {
			return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid column in mappings")
		}
		colList = append(colList, ident)
	}

	query := fmt.Sprintf(`SELECT %s FROM %s`, strings.Join(colList, ", "), tableName)

	rows, err := h.deps.DB.QueryContext(ctx, query)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "failed to query data")
	}
	defer rows.Close()

	headerRow, err := h.engine.Mapper.BuildHeaderRow(ctx, profile.ID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "failed to build header row")
	}

	x := excelize.NewFile()
	sheetName := strings.Title(scope)
	x.SetSheetName("Sheet1", sheetName)

	setCell := func(row, col int, val any) {
		cell, _ := excelize.CoordinatesToCellName(col, row)
		if err := x.SetCellValue(sheetName, cell, val); err != nil {
			logger.Warn("metadata.export.set_cell_failed", "row", row, "col", col, "err", err)
		}
	}

	for colIdx, v := range headerRow.Columns {
		setCell(1, colIdx, v)
	}

	dataRowIndex := 1

	for rows.Next() {
		dest := make([]any, len(colList))
		destPtrs := make([]any, len(colList))
		for i := range dest {
			destPtrs[i] = &dest[i]
		}

		if err := rows.Scan(destPtrs...); err != nil {
			logger.Error("metadata.export.scan_failed", "err", err)
			continue
		}

		// idx 0: id (bỏ qua)
		// idx 1: custom_fields JSONB
		rawCustom := dest[1]

		// parse custom_fields
		meta := map[string]any{}
		switch v := rawCustom.(type) {
		case []byte:
			if len(v) > 0 {
				_ = json.Unmarshal(v, &meta)
			}
		case string:
			if v != "" {
				_ = json.Unmarshal([]byte(v), &meta)
			}
		case nil:
			// no-op
		default:
			// no-op
		}

		core := map[string]any{}
		// colList: [ "id", "custom_fields", "name", "phone_number", ... ]
		for idx := 2; idx < len(colList); idx++ {
			colName := strings.Trim(colList[idx], `"`) // bỏ quote
			core[colName] = dest[idx]
		}

		entity := &service.EntityData{
			Core:     core,
			Metadata: meta,
			External: map[string]any{},
		}

		dataRowIndex++
		dataRow, err := h.engine.Mapper.BuildDataRow(ctx, profile.ID, entity)
		if err != nil {
			logger.Warn("metadata.export.build_data_row_failed", "err", err)
			continue
		}

		for colIdx, v := range dataRow.Columns {
			setCell(dataRowIndex, colIdx, v)
		}
	}

	if err := rows.Err(); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "error while reading rows")
	}

	filename := fmt.Sprintf("%s_export_%d.xlsx", scope, time.Now().Unix())
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	if err := x.Write(c.Response().BodyWriter()); err != nil {
		logger.Error("metadata.export.write_failed", "err", err)
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, "failed to write excel file")
	}

	return nil
}
