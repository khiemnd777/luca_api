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
	"github.com/khiemnd777/andy_api/modules/main/features/process/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type ProcessImportService interface {
	ImportFromExcel(ctx context.Context, deptID int, rows []model.ProcessExcelRow) (model.ProcessImportResult, error)
}

type processImportService struct {
	repo repository.ProcessImportRepository
	db   *sql.DB
}

func NewProcessImportService(repo repository.ProcessImportRepository, db *sql.DB) ProcessImportService {
	return &processImportService{repo: repo, db: db}
}

func (s *processImportService) ImportFromExcel(ctx context.Context, deptID int, rows []model.ProcessExcelRow) (model.ProcessImportResult, error) {
	result := model.ProcessImportResult{TotalRows: len(rows)}
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
		if row.Name == "" {
			result.Skipped++
			continue
		}

		_, created, err := s.repo.GetOrCreate(ctx, deptID, row.Name)
		if err != nil {
			return result, fmt.Errorf("row %d: cannot create process: %w", rowIndex, err)
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

	cache.InvalidateKeys(kProcessAll(deptID)...)
	return result, nil
}

func ParseProcessExcel(file io.Reader) ([]model.ProcessExcelRow, error) {
	x, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := x.Close(); err != nil {
			logger.Warn("process.import.close_excel_failed", "error", err)
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

	var out []model.ProcessExcelRow
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

		name := normalizeCell(getCell(cols, 0))
		if name == "" {
			continue
		}

		out = append(out, model.ProcessExcelRow{Name: name})
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
