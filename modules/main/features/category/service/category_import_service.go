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
	"github.com/khiemnd777/andy_api/modules/main/features/category/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type CategoryImportService interface {
	ImportFromExcel(ctx context.Context, deptID int, rows []model.CategoryExcelRow) (model.CategoryImportResult, error)
}

type categoryImportService struct {
	repo repository.CategoryImportRepository
	db   *sql.DB
}

func NewCategoryImportService(repo repository.CategoryImportRepository, db *sql.DB) CategoryImportService {
	return &categoryImportService{repo: repo, db: db}
}

func (s *categoryImportService) ImportFromExcel(ctx context.Context, deptID int, rows []model.CategoryExcelRow) (model.CategoryImportResult, error) {
	result := model.CategoryImportResult{TotalRows: len(rows)}
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
		rowIndex := idx + 2 // +1 header, +1 for 1-based
		addedThisRow := false

		lv1ID, created, err := s.repo.GetOrCreateLV1(ctx, deptID, row.LV1)
		if err != nil {
			return result, fmt.Errorf("row %d: cannot create lv1: %w", rowIndex, err)
		}
		if created {
			result.AddedLV1++
			addedThisRow = true
		}

		if row.LV2 == "" {
			if !addedThisRow {
				result.Skipped++
			}
			continue
		}

		lv2ID, created, err := s.repo.GetOrCreateLV2(ctx, deptID, lv1ID, row.LV1, row.LV2)
		if err != nil {
			return result, fmt.Errorf("row %d: cannot create lv2: %w", rowIndex, err)
		}
		if created {
			result.AddedLV2++
			addedThisRow = true
		}

		if row.LV3 == "" {
			if !addedThisRow {
				result.Skipped++
			}
			continue
		}

		_, created, err = s.repo.GetOrCreateLV3(ctx, deptID, lv1ID, lv2ID, row.LV1, row.LV2, row.LV3)
		if err != nil {
			return result, fmt.Errorf("row %d: cannot create lv3: %w", rowIndex, err)
		}
		if created {
			result.AddedLV3++
			addedThisRow = true
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

		if lv1 == "" {
			continue
		}

		out = append(out, model.CategoryExcelRow{
			LV1: lv1,
			LV2: lv2,
			LV3: lv3,
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
