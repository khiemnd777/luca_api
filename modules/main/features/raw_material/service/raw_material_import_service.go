package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/raw_material/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type RawMaterialImportService interface {
	ImportFromExcel(ctx context.Context, deptID int, rows []model.RawMaterialExcelRow) (model.RawMaterialImportResult, error)
}

type rawMaterialImportService struct {
	repo repository.RawMaterialImportRepository
	db   *sql.DB
}

func NewRawMaterialImportService(repo repository.RawMaterialImportRepository, db *sql.DB) RawMaterialImportService {
	return &rawMaterialImportService{repo: repo, db: db}
}

func (s *rawMaterialImportService) ImportFromExcel(ctx context.Context, deptID int, rows []model.RawMaterialExcelRow) (model.RawMaterialImportResult, error) {
	result := model.RawMaterialImportResult{TotalRows: len(rows)}
	if len(rows) == 0 {
		return result, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
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
		rowIndex := idx + 2
		if row.CategoryName == "" {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: missing category name", rowIndex))
			continue
		}
		if row.Name == "" {
			result.Skipped++
			continue
		}

		categoryID, categoryName, err := s.repo.GetCategoryByName(ctx, deptID, row.CategoryName)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				result.Skipped++
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: category not found: %s", rowIndex, row.CategoryName))
				continue
			}
			return result, fmt.Errorf("row %d: lookup category failed: %w", rowIndex, err)
		}

		_, created, err := s.repo.GetOrCreateRawMaterial(ctx, deptID, categoryID, categoryName, row.Name)
		if err != nil {
			return result, fmt.Errorf("row %d: cannot create raw material: %w", rowIndex, err)
		}
		if created {
			result.Added++
		} else {
			result.Skipped++
		}
	}

	if err := tx.Commit(); err != nil {
		return result, err
	}
	committed = true

	cache.InvalidateKeys(kRawMaterialAll(deptID)...)
	return result, nil
}

func ParseRawMaterialExcel(file io.Reader) ([]model.RawMaterialExcelRow, error) {
	x, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := x.Close(); err != nil {
			logger.Warn("raw_material.import.close_excel_failed", "error", err)
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

	var out []model.RawMaterialExcelRow
	rowIndex := 0
	lastCategory := ""
	for rows.Next() {
		rowIndex++
		cols, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("row %d: cannot read columns: %w", rowIndex, err)
		}
		if rowIndex == 1 {
			continue
		}

		cat := normalizeCell(getCell(cols, 0))
		name := normalizeCell(getCell(cols, 1))

		if cat == "" {
			cat = lastCategory
		}

		if cat == "" && name == "" {
			continue
		}

		if cat != "" {
			lastCategory = cat
		}

		if name == "" {
			out = append(out, model.RawMaterialExcelRow{
				CategoryName: cat,
				Name:         "",
			})
			continue
		}

		out = append(out, model.RawMaterialExcelRow{
			CategoryName: cat,
			Name:         name,
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
