package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/khiemnd777/andy_api/modules/metadata/model"
	"github.com/khiemnd777/andy_api/shared/logger"
)

var identRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type ImportEngine struct {
	DB     *sql.DB
	Mapper *ImportFieldMappingService
}

func NewImportEngine(db *sql.DB, mapper *ImportFieldMappingService) *ImportEngine {
	return &ImportEngine{DB: db, Mapper: mapper}
}

func (e *ImportEngine) SafeIdent(name string) (string, error) {
	name = strings.TrimSpace(name)
	if !identRegex.MatchString(name) {
		return "", fmt.Errorf("unsafe identifier: %s", name)
	}
	// quote identifier cho Postgres
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`, nil
}

func findPivotMapping(
	profile *model.ImportFieldProfile,
	mappings []model.ImportFieldMapping,
	row *model.MappedRow,
) (*model.ImportFieldMapping, any, error) {
	if profile.PivotField == nil || strings.TrimSpace(*profile.PivotField) == "" {
		return nil, nil, fmt.Errorf("pivot_field is not set for profile %d", profile.ID)
	}
	pivotPath := strings.TrimSpace(*profile.PivotField)

	var pivotMap *model.ImportFieldMapping
	for i := range mappings {
		if mappings[i].InternalPath == pivotPath {
			pivotMap = &mappings[i]
			break
		}
	}
	if pivotMap == nil {
		return nil, nil, fmt.Errorf("pivot_field %s not found in mappings", pivotPath)
	}

	var val any
	switch pivotMap.InternalKind {
	case "core":
		val = row.CoreFields[pivotMap.InternalPath]
	case "metadata":
		val = row.MetadataFields[pivotMap.InternalPath]
	case "external":
		val = row.ExternalFields[pivotMap.InternalPath]
	default:
		return nil, nil, fmt.Errorf("unsupported pivot kind: %s", pivotMap.InternalKind)
	}

	// pivot trống thì bỏ qua (tuỳ logic, ở đây coi là error nhẹ)
	if val == nil || fmt.Sprint(val) == "" {
		return nil, nil, fmt.Errorf("pivot value is empty")
	}

	return pivotMap, val, nil
}

func (e *ImportEngine) findExistingID(
	ctx context.Context,
	scope string,
	pivotMap *model.ImportFieldMapping,
	pivotVal any,
) (int64, error) {
	table, err := e.SafeIdent(scope)
	if err != nil {
		return 0, err
	}

	var query string
	var args []any

	switch pivotMap.InternalKind {
	case "core":
		col, err := e.SafeIdent(pivotMap.InternalPath)
		if err != nil {
			return 0, err
		}
		query = fmt.Sprintf(`SELECT id FROM %s WHERE %s = $1 LIMIT 1`, table, col)
		args = []any{pivotVal}

	case "metadata":
		// SELECT id FROM clinics WHERE custom_fields->>'tax_code' = $1
		query = fmt.Sprintf(
			`SELECT id FROM %s WHERE custom_fields->>'%s' = $1 LIMIT 1`,
			table,
			pivotMap.InternalPath,
		)
		args = []any{fmt.Sprint(pivotVal)}

	default:
		return 0, fmt.Errorf("unsupported pivot kind: %s", pivotMap.InternalKind)
	}

	var id int64
	err = e.DB.QueryRowContext(ctx, query, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

func buildCoreAndCustom(row *model.MappedRow) (map[string]any, map[string]any) {
	core := map[string]any{}
	for k, v := range row.CoreFields {
		core[k] = v
	}
	meta := map[string]any{}
	for k, v := range row.MetadataFields {
		meta[k] = v
	}
	return core, meta
}

func (e *ImportEngine) UpsertMappedRow(
	ctx context.Context,
	scope string,
	profile *model.ImportFieldProfile,
	mappings []model.ImportFieldMapping,
	row *model.MappedRow,
) (created bool, err error) {
	pivotMap, pivotVal, err := findPivotMapping(profile, mappings, row)
	if err != nil {
		// tuỳ bạn coi là error hay skip; ở đây log + skip
		logger.Warn("metadata.import.pivot_error", "err", err, "scope", scope)
		return false, err
	}

	table, err := e.SafeIdent(scope)
	if err != nil {
		return false, err
	}

	// tìm id hiện có
	existingID, err := e.findExistingID(ctx, scope, pivotMap, pivotVal)
	if err != nil {
		return false, err
	}

	core, meta := buildCoreAndCustom(row)
	metaJSON, _ := json.Marshal(meta)

	if existingID == 0 {
		// INSERT
		cols := []string{}
		args := []any{}
		placeholders := []string{}
		i := 1

		for k, v := range core {
			col, err := e.SafeIdent(k)
			if err != nil {
				return false, err
			}
			cols = append(cols, col)
			args = append(args, v)
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
			i++
		}

		// custom_fields
		cols = append(cols, `"custom_fields"`)
		args = append(args, string(metaJSON))
		placeholders = append(placeholders, fmt.Sprintf("$%d::jsonb", i))

		query := fmt.Sprintf(
			`INSERT INTO %s (%s) VALUES (%s)`,
			table,
			strings.Join(cols, ", "),
			strings.Join(placeholders, ", "),
		)

		if _, err := e.DB.ExecContext(ctx, query, args...); err != nil {
			return false, err
		}
		return true, nil
	}

	// UPDATE
	setClauses := []string{}
	args := []any{}
	i := 1

	for k, v := range core {
		col, err := e.SafeIdent(k)
		if err != nil {
			return false, err
		}
		setClauses = append(setClauses, fmt.Sprintf(`%s = $%d`, col, i))
		args = append(args, v)
		i++
	}

	// merge custom_fields
	setClauses = append(setClauses, fmt.Sprintf(`custom_fields = COALESCE(custom_fields, '{}'::jsonb) || $%d::jsonb`, i))
	args = append(args, string(metaJSON))
	i++

	query := fmt.Sprintf(
		`UPDATE %s SET %s WHERE id = $%d`,
		table,
		strings.Join(setClauses, ", "),
		i,
	)
	args = append(args, existingID)

	if _, err := e.DB.ExecContext(ctx, query, args...); err != nil {
		return false, err
	}
	return false, nil
}
