package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/category/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/logger"
	collectionutils "github.com/khiemnd777/andy_api/shared/metadata/collection"
)

type CategoryImportService interface {
	ImportFromExcel(ctx context.Context, deptID int, rows []model.CategoryExcelRow) (model.CategoryImportResult, error)
}

type categoryImportService struct {
	repo repository.CategoryImportRepository
	db   *generated.Client
}

func NewCategoryImportService(repo repository.CategoryImportRepository, db *generated.Client) CategoryImportService {
	return &categoryImportService{repo: repo, db: db}
}

func (s *categoryImportService) ImportFromExcel(ctx context.Context, deptID int, rows []model.CategoryExcelRow) (model.CategoryImportResult, error) {
	result := model.CategoryImportResult{TotalRows: len(rows)}
	if len(rows) == 0 {
		return result, nil
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return result, err
	}
	ctx = repository.WithTx(ctx, tx)
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	for idx, row := range rows {
		rowIndex := idx + 2 // +1 header, +1 for 1-based
		addedThisRow := false

		lv1ID, created, err := s.repo.GetOrCreateLV1(ctx, deptID, row.LV1)
		if err != nil {
			return result, fmt.Errorf("row %d: cannot create lv1: %w", rowIndex, err)
		}
		if created {
			result.AddedLV1++
			addedThisRow = true
			if err := s.ensureCollections(ctx, tx, deptID, lv1ID); err != nil {
				return result, fmt.Errorf("row %d: cannot update lv1 collections: %w", rowIndex, err)
			}
		}

		lv2ID := 0
		if row.LV2 != "" {
			lv2ID, created, err = s.repo.GetOrCreateLV2(ctx, deptID, lv1ID, row.LV1, row.LV2)
			if err != nil {
				return result, fmt.Errorf("row %d: cannot create lv2: %w", rowIndex, err)
			}
			if created {
				result.AddedLV2++
				addedThisRow = true
				if err := collectionutils.UpsertAncestorCollections(ctx, tx, categoryTreeCfg, lv2ID); err != nil {
					return result, fmt.Errorf("row %d: cannot update ancestor collections for lv2: %w", rowIndex, err)
				}
			}
		}

		lv3ID := 0
		if row.LV3 != "" && lv2ID > 0 {
			lv3ID, created, err = s.repo.GetOrCreateLV3(ctx, deptID, lv1ID, lv2ID, row.LV1, row.LV2, row.LV3)
			if err != nil {
				return result, fmt.Errorf("row %d: cannot create lv3: %w", rowIndex, err)
			}
			if created {
				result.AddedLV3++
				addedThisRow = true
				if err := collectionutils.UpsertAncestorCollections(ctx, tx, categoryTreeCfg, lv3ID); err != nil {
					return result, fmt.Errorf("row %d: cannot update ancestor collections for lv3: %w", rowIndex, err)
				}
			}
		}

		if err := s.applyFieldsForRow(ctx, tx, deptID, row, lv1ID, lv2ID, lv3ID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", rowIndex, err))
		}

		if !addedThisRow {
			result.Skipped++
		}
	}

	if err := tx.Commit(); err != nil {
		return result, err
	}
	committed = true

	cache.InvalidateKeys(kCategoryAll(deptID)...)

	return result, nil
}

var categoryTreeCfg = collectionutils.TreeConfig{
	TableName:        "categories",
	IDColumn:         "id",
	ParentIDColumn:   "parent_id",
	ShowIfFieldName:  "categoryId",
	CollectionGroup:  "category",
	CollectionPrefix: "category",
}

func (s *categoryImportService) ensureCollections(ctx context.Context, tx *generated.Tx, deptID int, categoryID int) error {
	node, err := s.repo.GetTreeNode(ctx, deptID, categoryID)
	if err != nil {
		return err
	}
	if err := collectionutils.UpsertCollectionForNode(ctx, tx, categoryTreeCfg, node, nil); err != nil {
		return err
	}
	return collectionutils.UpsertAncestorCollections(ctx, tx, categoryTreeCfg, categoryID)
}

func (s *categoryImportService) applyFieldsForRow(ctx context.Context, tx *generated.Tx, deptID int, row model.CategoryExcelRow, lv1ID, lv2ID, lv3ID int) error {
	fields, fieldErrs := buildCategoryFieldSpecs(row)
	if len(fields) == 0 {
		if len(fieldErrs) > 0 {
			return fmt.Errorf("field parse error: %s", strings.Join(fieldErrs, "; "))
		}
		return nil
	}

	targetID := lv1ID

	if err := s.ensureCollections(ctx, tx, deptID, targetID); err != nil {
		return err
	}

	collectionID, err := s.repo.GetCollectionID(ctx, deptID, targetID)
	if err != nil {
		return err
	}
	if collectionID == nil {
		return fmt.Errorf("collection_id not found for category %d", targetID)
	}

	_, err = s.repo.UpsertFields(ctx, *collectionID, fields)
	if err != nil {
		return err
	}
	if len(fieldErrs) > 0 {
		return fmt.Errorf("field parse error: %s", strings.Join(fieldErrs, "; "))
	}
	return nil
}

func ParseCategoryExcel(file io.Reader) ([]model.CategoryExcelRow, error) {
	x, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := x.Close(); err != nil {
			logger.Warn("category.import.close_excel_failed", "error", err)
		}
	}()

	sheet := x.GetSheetName(0)
	if sheet == "" {
		return nil, errors.New("empty sheet")
	}

	rows, err := x.Rows(sheet)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var out []model.CategoryExcelRow
	rowIndex := 0
	lastLV1 := ""
	lastLV2 := ""
	for rows.Next() {
		rowIndex++
		cols, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("row %d: cannot read columns: %w", rowIndex, err)
		}
		if rowIndex == 1 {
			continue
		}

		lv1 := normalizeCell(getCell(cols, 0))
		lv2 := normalizeCell(getCell(cols, 1))
		lv3 := normalizeCell(getCell(cols, 2))
		field1 := normalizeCell(getCell(cols, 3))
		field3 := normalizeCell(getCell(cols, 4))
		field4 := normalizeCell(getCell(cols, 5))
		field5 := normalizeCell(getCell(cols, 6))

		if lv1 == "" {
			lv1 = lastLV1
		}
		if lv2 == "" {
			lv2 = lastLV2
		}

		if lv1 == "" && lv2 == "" && lv3 == "" && field1 == "" && field3 == "" && field4 == "" && field5 == "" {
			continue
		}

		if lv1 != "" {
			lastLV1 = lv1
		}
		if lv2 != "" {
			lastLV2 = lv2
		}

		out = append(out, model.CategoryExcelRow{
			LV1:    lv1,
			LV2:    lv2,
			LV3:    lv3,
			Field1: field1,
			Field3: field3,
			Field4: field4,
			Field5: field5,
		})
	}

	return out, nil
}

func getCell(cols []string, idx int) string {
	if idx < len(cols) {
		return cols[idx]
	}
	return ""
}

func normalizeCell(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.Join(strings.Fields(value), " ")
}

func buildCategoryFieldSpecs(row model.CategoryExcelRow) ([]repository.CategoryFieldSpec, []string) {
	type fieldCell struct {
		Value      string
		OrderIndex int
	}

	cells := []fieldCell{
		{Value: row.Field1, OrderIndex: 1},
		{Value: row.Field3, OrderIndex: 3},
		{Value: row.Field4, OrderIndex: 4},
		{Value: row.Field5, OrderIndex: 5},
	}

	var out []repository.CategoryFieldSpec
	var errs []string
	for _, cell := range cells {
		if cell.Value == "" {
			continue
		}
		spec, err := parseFieldSpec(cell.Value)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		spec.OrderIndex = cell.OrderIndex
		out = append(out, spec)
	}
	return out, errs
}

func parseFieldSpec(raw string) (repository.CategoryFieldSpec, error) {
	var spec repository.CategoryFieldSpec
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return spec, errors.New("empty field spec")
	}

	open := strings.Index(raw, "{")
	close := strings.LastIndex(raw, "}")
	if open < 0 || close < 0 || close <= open {
		return spec, fmt.Errorf("invalid field spec format")
	}

	prefix := strings.TrimSpace(raw[:open])
	body := raw[open : close+1]

	jsonBody, err := normalizeFieldJSON(body)
	if err != nil {
		return spec, err
	}

	var obj map[string]any
	if err := json.Unmarshal([]byte(jsonBody), &obj); err != nil {
		return spec, err
	}

	name, _ := obj["name"].(string)
	label, _ := obj["label"].(string)
	typ, _ := obj["type"].(string)
	visibility, _ := obj["visibility"].(string)

	if label == "" && prefix != "" {
		label = prefix
	}

	spec.Name = name
	spec.Label = label
	spec.Type = typ
	if spec.Type == "" {
		spec.Type = "text"
	}
	if visibility == "" {
		visibility = "public"
	}
	spec.Visibility = visibility

	if v, ok := obj["required"].(bool); ok {
		spec.Required = v
	}
	if v, ok := obj["unique"].(bool); ok {
		spec.Unique = v
	}
	if v, ok := obj["table"].(bool); ok {
		spec.Table = v
	}
	if v, ok := obj["form"].(bool); ok {
		spec.Form = v
	}
	if v, ok := obj["search"].(bool); ok {
		spec.Search = v
	}
	if v, ok := obj["tag"].(string); ok && v != "" {
		spec.Tag = &v
	}

	if rel, ok := obj["relation"]; ok {
		if b, err := json.Marshal(rel); err == nil {
			s := string(b)
			spec.Relation = &s
		}
	}
	if opts, ok := obj["options"]; ok {
		if b, err := json.Marshal(opts); err == nil {
			s := string(b)
			spec.Options = &s
		}
	}
	if def, ok := obj["default_value"]; ok {
		if b, err := json.Marshal(def); err == nil {
			s := string(b)
			spec.DefaultValue = &s
		}
	}

	if spec.Name == "" || spec.Label == "" {
		return spec, fmt.Errorf("field spec missing name or label")
	}
	return spec, nil
}

func normalizeFieldJSON(body string) (string, error) {
	var out strings.Builder
	inString := false
	escaped := false
	runes := []rune(body)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if ch == '\\' && !escaped {
			escaped = true
			out.WriteRune(ch)
			continue
		}
		if ch == '"' && !escaped {
			inString = !inString
			out.WriteRune(ch)
			continue
		}
		escaped = false

		if !inString && isIdentStart(ch) {
			start := i
			j := i + 1
			for j < len(runes) && isIdentPart(runes[j]) {
				j++
			}
			k := j
			for k < len(runes) && (runes[k] == ' ' || runes[k] == '\t' || runes[k] == '\n' || runes[k] == '\r') {
				k++
			}
			if k < len(runes) && runes[k] == ':' {
				key := string(runes[start:j])
				out.WriteRune('"')
				out.WriteString(key)
				out.WriteString(`":`)
				i = k
				continue
			}
		}

		out.WriteRune(ch)
	}
	return out.String(), nil
}

func isIdentStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentPart(ch rune) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}
