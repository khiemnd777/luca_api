package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/xuri/excelize/v2"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/product/repository"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type ProductImportService interface {
	ImportFromExcel(ctx context.Context, deptID int, rows []model.ProductExcelRow) (model.ProductImportResult, error)
}

type productImportService struct {
	repo       repository.ProductImportRepository
	productSvc ProductService
}

func NewProductImportService(repo repository.ProductImportRepository, productSvc ProductService) ProductImportService {
	return &productImportService{repo: repo, productSvc: productSvc}
}

func (s *productImportService) ImportFromExcel(ctx context.Context, deptID int, rows []model.ProductExcelRow) (model.ProductImportResult, error) {
	result := model.ProductImportResult{TotalRows: len(rows)}
	if len(rows) == 0 {
		return result, nil
	}

	for idx, row := range rows {
		rowIndex := idx + 2

		if row.Code == "" {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: missing product code", rowIndex))
			continue
		}
		if row.Name == "" {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: missing product name", rowIndex))
			continue
		}
		if row.CategoryLV1 == "" {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: missing category level 1", rowIndex))
			continue
		}

		categoryID, categoryName, err := s.repo.ResolveCategoryBranch(ctx, deptID, row.CategoryLV1, row.CategoryLV2, row.CategoryLV3)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				result.Skipped++
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: category not found for branch [%s > %s > %s]", rowIndex, row.CategoryLV1, row.CategoryLV2, row.CategoryLV3))
				continue
			}
			return result, fmt.Errorf("row %d: cannot resolve category: %w", rowIndex, err)
		}
		lv1ID, lv1Name, err := s.repo.ResolveCategoryLV1(ctx, deptID, row.CategoryLV1)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				result.Skipped++
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: category level 1 not found [%s]", rowIndex, row.CategoryLV1))
				continue
			}
			return result, fmt.Errorf("row %d: cannot resolve category level 1: %w", rowIndex, err)
		}

		var brandIDs []int
		if row.HasBrandField {
			brandIDs, err = s.resolveCategoryRefIDs(ctx, deptID, lv1ID, lv1Name, row.BrandNames, s.repo.GetOrCreateBrandName)
			if err != nil {
				return result, fmt.Errorf("row %d: cannot resolve brand names: %w", rowIndex, err)
			}
		}

		var rawMaterialIDs []int
		if row.HasRawMaterialField {
			rawMaterialIDs, err = s.resolveCategoryRefIDs(ctx, deptID, lv1ID, lv1Name, row.RawMaterialNames, s.repo.GetOrCreateRawMaterial)
			if err != nil {
				return result, fmt.Errorf("row %d: cannot resolve raw materials: %w", rowIndex, err)
			}
		}

		var techniqueIDs []int
		if row.HasTechniqueField {
			techniqueIDs, err = s.resolveCategoryRefIDs(ctx, deptID, lv1ID, lv1Name, row.TechniqueNames, s.repo.GetOrCreateTechnique)
			if err != nil {
				return result, fmt.Errorf("row %d: cannot resolve techniques: %w", rowIndex, err)
			}
		}

		var restorationTypeIDs []int
		if row.HasRestorationTypeField {
			restorationTypeIDs, err = s.resolveCategoryRefIDs(ctx, deptID, lv1ID, lv1Name, row.RestorationTypeNames, s.repo.GetOrCreateRestorationType)
			if err != nil {
				return result, fmt.Errorf("row %d: cannot resolve restoration types: %w", rowIndex, err)
			}
		}

		var processIDs []int
		if row.HasProcessField {
			processIDs, err = s.resolveProcessIDs(ctx, deptID, row.ProcessNames)
			if err != nil {
				return result, fmt.Errorf("row %d: cannot resolve processes: %w", rowIndex, err)
			}
		}

		existing, err := s.repo.FindProductByCode(ctx, deptID, row.Code)
		if err != nil {
			return result, fmt.Errorf("row %d: cannot lookup product code %s: %w", rowIndex, row.Code, err)
		}

		payload := &model.ProductUpsertDTO{DTO: model.ProductDTO{
			Code:         utils.Ptr(row.Code),
			Name:         utils.Ptr(row.Name),
			CategoryID:   utils.Ptr(categoryID),
			CategoryName: utils.Ptr(categoryName),
		}}

		price, err := parseImportedPrice(row.RetailPrice)
		if err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: invalid retail price %q", rowIndex, row.RetailPrice))
			continue
		}
		if price != nil {
			payload.DTO.RetailPrice = price
			payload.DTO.CostPrice = price
		}

		if row.HasBrandField {
			payload.DTO.BrandNameIDs = brandIDs
		}
		if row.HasRawMaterialField {
			payload.DTO.RawMaterialIDs = rawMaterialIDs
		}
		if row.HasTechniqueField {
			payload.DTO.TechniqueIDs = techniqueIDs
		}
		if row.HasRestorationTypeField {
			payload.DTO.RestorationTypeIDs = restorationTypeIDs
		}
		if row.HasProcessField {
			payload.DTO.ProcessIDs = processIDs
		}

		if existing != nil {
			payload.DTO.ID = existing.ID
			payload.DTO.TemplateID = existing.TemplateID
			if _, err := s.productSvc.Update(ctx, deptID, payload); err != nil {
				return result, fmt.Errorf("row %d: cannot update product code %s: %w", rowIndex, row.Code, err)
			}
			result.Updated++
			continue
		}

		if _, err := s.productSvc.Create(ctx, deptID, payload); err != nil {
			return result, fmt.Errorf("row %d: cannot create product code %s: %w", rowIndex, row.Code, err)
		}
		result.Added++
	}

	return result, nil
}

type categoryRefUpserter func(ctx context.Context, deptID int, categoryID int, categoryName, name string) (id int, created bool, err error)

func (s *productImportService) resolveCategoryRefIDs(
	ctx context.Context,
	deptID int,
	categoryID int,
	categoryName string,
	raw string,
	upsert categoryRefUpserter,
) ([]int, error) {
	names := splitMultiNames(raw)
	if len(names) == 0 {
		return []int{}, nil
	}

	out := make([]int, 0, len(names))
	for _, name := range names {
		id, _, err := upsert(ctx, deptID, categoryID, categoryName, name)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return dedupInts(out), nil
}

func (s *productImportService) resolveProcessIDs(ctx context.Context, deptID int, raw string) ([]int, error) {
	names := splitProcessNames(raw)
	if len(names) == 0 {
		return []int{}, nil
	}

	out := make([]int, 0, len(names))
	for _, name := range names {
		id, _, err := s.repo.GetOrCreateProcess(ctx, deptID, name)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return dedupInts(out), nil
}

func ParseProductExcel(file io.Reader) ([]model.ProductExcelRow, error) {
	x, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := x.Close(); err != nil {
			logger.Warn("product.import.close_excel_failed", "error", err)
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

	var out []model.ProductExcelRow
	rowIndex := 0
	lastLV1 := ""
	lastLV2 := ""
	colIndex := productExcelColumnIndex{}

	for rows.Next() {
		rowIndex++
		cols, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("row %d: cannot read columns: %w", rowIndex, err)
		}

		if rowIndex == 1 {
			colIndex = detectProductColumns(cols)
			continue
		}

		row := model.ProductExcelRow{
			Code:                    normalizeCell(readByIndex(cols, colIndex.Code, 0)),
			Name:                    normalizeCell(readByIndex(cols, colIndex.Name, 1)),
			CategoryLV1:             normalizeCell(readByIndex(cols, colIndex.CategoryLV1, 2)),
			CategoryLV2:             normalizeCell(readByIndex(cols, colIndex.CategoryLV2, 3)),
			CategoryLV3:             normalizeCell(readByIndex(cols, colIndex.CategoryLV3, 4)),
			BrandNames:              normalizeCell(readByIndex(cols, colIndex.BrandNames, 5)),
			ProcessNames:            normalizeCell(readByIndex(cols, colIndex.ProcessNames, 6)),
			RetailPrice:             normalizeCell(readByIndex(cols, colIndex.RetailPrice, 7)),
			RawMaterialNames:        normalizeCell(readByIndex(cols, colIndex.RawMaterialNames, 8)),
			TechniqueNames:          normalizeCell(readByIndex(cols, colIndex.TechniqueNames, 9)),
			RestorationTypeNames:    normalizeCell(readByIndex(cols, colIndex.RestorationTypeNames, 10)),
			HasBrandField:           colIndex.BrandNames >= 0 || len(cols) > 5,
			HasProcessField:         colIndex.ProcessNames >= 0 || len(cols) > 6,
			HasRawMaterialField:     colIndex.RawMaterialNames >= 0 || len(cols) > 8,
			HasTechniqueField:       colIndex.TechniqueNames >= 0 || len(cols) > 9,
			HasRestorationTypeField: colIndex.RestorationTypeNames >= 0 || len(cols) > 10,
		}

		if row.CategoryLV1 == "" {
			row.CategoryLV1 = lastLV1
		}
		if row.CategoryLV2 == "" {
			row.CategoryLV2 = lastLV2
		}

		if row.CategoryLV1 != "" {
			lastLV1 = row.CategoryLV1
		}
		if row.CategoryLV2 != "" {
			lastLV2 = row.CategoryLV2
		}

		if row.Code == "" && row.Name == "" && row.CategoryLV1 == "" && row.CategoryLV2 == "" && row.CategoryLV3 == "" && row.BrandNames == "" && row.ProcessNames == "" && row.RawMaterialNames == "" && row.TechniqueNames == "" && row.RestorationTypeNames == "" {
			continue
		}

		out = append(out, row)
	}

	return out, nil
}

type productExcelColumnIndex struct {
	Code                 int
	Name                 int
	CategoryLV1          int
	CategoryLV2          int
	CategoryLV3          int
	BrandNames           int
	RawMaterialNames     int
	TechniqueNames       int
	RestorationTypeNames int
	ProcessNames         int
	RetailPrice          int
}

func detectProductColumns(headers []string) productExcelColumnIndex {
	idx := productExcelColumnIndex{
		Code:                 -1,
		Name:                 -1,
		CategoryLV1:          -1,
		CategoryLV2:          -1,
		CategoryLV3:          -1,
		BrandNames:           -1,
		RawMaterialNames:     -1,
		TechniqueNames:       -1,
		RestorationTypeNames: -1,
		ProcessNames:         -1,
		RetailPrice:          -1,
	}

	for i, h := range headers {
		key := detectHeaderKey(h)
		switch key {
		case "code":
			idx.Code = i
		case "name":
			idx.Name = i
		case "category_lv1":
			idx.CategoryLV1 = i
		case "category_lv2":
			idx.CategoryLV2 = i
		case "category_lv3":
			idx.CategoryLV3 = i
		case "brand":
			idx.BrandNames = i
		case "raw_material":
			idx.RawMaterialNames = i
		case "technique":
			idx.TechniqueNames = i
		case "restoration_type":
			idx.RestorationTypeNames = i
		case "process":
			idx.ProcessNames = i
		case "retail_price":
			idx.RetailPrice = i
		}
	}

	return idx
}

func detectHeaderKey(value string) string {
	n := normalizeHeader(value)
	if n == "" {
		return ""
	}

	has := func(token string) bool {
		return strings.Contains(n, token)
	}

	switch {
	case has("code") || has("masanpham") || has("mãsảnphẩm"):
		return "code"
	case has("name") || has("motasanpham") || has("môtảsảnphẩm"):
		return "name"
	case has("categorylv1") || has("lv1") || has("phuchinh") || has("phụchình"):
		return "category_lv1"
	case has("categorylv2") || has("lv2") || has("phanloai") || has("phânloại"):
		return "category_lv2"
	case has("categorylv3") || has("lv3") || has("phuongphap") || has("phươngpháp"):
		return "category_lv3"
	case has("brand") || has("thuonghieu") || has("thươnghiệu"):
		return "brand"
	case has("rawmaterial") || has("vatlieu") || has("vậtliệu"):
		return "raw_material"
	case has("technique") || has("congnghe") || has("côngnghệ"):
		return "technique"
	case has("restoration") || has("kieuphuchinh") || has("kiểuphụchình"):
		return "restoration_type"
	case has("process") || has("congdoan") || has("côngđoạn"):
		return "process"
	case has("retailprice") || has("banggia") || has("bảnggiá") || has("giaban") || has("giábán"):
		return "retail_price"
	default:
		return ""
	}
}

func readByIndex(cols []string, idx int, fallback int) string {
	if idx >= 0 && idx < len(cols) {
		return cols[idx]
	}
	if fallback >= 0 && fallback < len(cols) {
		return cols[fallback]
	}
	return ""
}

func normalizeHeader(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	value = strings.ReplaceAll(value, "<br>", " ")
	value = strings.ReplaceAll(value, "\n", " ")

	var b strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizeCell(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.Join(strings.Fields(value), " ")
}

func splitMultiNames(value string) []string {
	if normalizeCell(value) == "" {
		return nil
	}

	s := strings.NewReplacer(";", ",", "|", ",", "/", ",", "–", ",", "—", ",", "-", ",").Replace(value)
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		v := normalizeCell(part)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return dedupStrings(out)
}

func splitProcessNames(value string) []string {
	if normalizeCell(value) == "" {
		return nil
	}

	s := strings.NewReplacer("–", "-", "—", "-", "|", "-", ";", "-").Replace(value)
	parts := strings.Split(s, "-")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		v := normalizeCell(part)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return dedupStrings(out)
}

func dedupStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(item)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func dedupInts(items []int) []int {
	if len(items) == 0 {
		return []int{}
	}
	seen := make(map[int]struct{}, len(items))
	out := make([]int, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func parseImportedPrice(raw string) (*float64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, nil
	}

	// Keep only digits and minus sign.
	var b strings.Builder
	for i, r := range s {
		if unicode.IsDigit(r) || (r == '-' && i == 0) {
			b.WriteRune(r)
		}
	}

	s = b.String()
	if s == "" || s == "-" {
		return nil, fmt.Errorf("invalid number")
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}

	return &v, nil
}
