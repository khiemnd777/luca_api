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
	"github.com/khiemnd777/andy_api/modules/main/features/section/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/logger"
)

type SectionImportService interface {
	ImportFromExcel(ctx context.Context, deptID int, rows []model.SectionExcelRow) (model.SectionImportResult, error)
}

type sectionImportService struct {
	repo repository.SectionImportRepository
	db   *sql.DB
}

func NewSectionImportService(repo repository.SectionImportRepository, db *sql.DB) SectionImportService {
	return &sectionImportService{repo: repo, db: db}
}

func (s *sectionImportService) ImportFromExcel(ctx context.Context, deptID int, rows []model.SectionExcelRow) (model.SectionImportResult, error) {
	result := model.SectionImportResult{TotalRows: len(rows)}
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

	sectionIDsTouched := map[int]bool{}
	sectionOrder := map[int]int{}
	for idx, row := range rows {
		rowIndex := idx + 2
		if row.Name == "" {
			result.Skipped++
			continue
		}

		sectionID, created, err := s.repo.GetOrCreateSection(ctx, deptID, row.Name, nullableString(row.Color))
		if err != nil {
			return result, fmt.Errorf("row %d: cannot create section: %w", rowIndex, err)
		}
		if created {
			result.Added++
		}

		sectionIDsTouched[sectionID] = true

		if row.ProcessName == "" {
			result.Skipped++
			continue
		}

		processID, processName, err := s.repo.GetProcessByName(ctx, deptID, row.ProcessName)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				result.Skipped++
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: process not found: %s", rowIndex, row.ProcessName))
				continue
			}
			return result, fmt.Errorf("row %d: lookup process failed: %w", rowIndex, err)
		}

		sectionOrder[sectionID]++
		_, err = s.repo.UpsertSectionProcess(ctx, sectionID, row.Name, processID, processName, nullableString(row.Color), sectionOrder[sectionID])
		if err != nil {
			return result, fmt.Errorf("row %d: upsert section process failed: %w", rowIndex, err)
		}

		if err := s.repo.UpdateProcessSectionCache(ctx, deptID, processID, sectionID, row.Name, nullableString(row.Color)); err != nil {
			return result, fmt.Errorf("row %d: update process cache failed: %w", rowIndex, err)
		}
	}

	for sectionID := range sectionIDsTouched {
		if err := s.repo.UpdateSectionProcessNames(ctx, sectionID); err != nil {
			return result, err
		}
	}

	if err := tx.Commit(); err != nil {
		return result, err
	}
	committed = true

	cache.InvalidateKeys(kSectionAll(deptID)...)
	cache.InvalidateKeys(kProcessAll(deptID)...)
	return result, nil
}

func ParseSectionExcel(file io.Reader) ([]model.SectionExcelRow, error) {
	x, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := x.Close(); err != nil {
			logger.Warn("section.import.close_excel_failed", "error", err)
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

	var out []model.SectionExcelRow
	rowIndex := 0
	lastName := ""
	lastColor := ""
	for rows.Next() {
		rowIndex++
		cols, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("row %d: cannot read columns: %w", rowIndex, err)
		}
		if rowIndex == 1 {
			continue
		}

		sectionCell := normalizeCell(getCell(cols, 0))
		processName := normalizeCell(getCell(cols, 1))

		name := sectionCell
		color := ""
		if sectionCell != "" {
			name, color = parseSectionCell(sectionCell)
			lastName = name
			lastColor = color
		} else {
			name = lastName
			color = lastColor
		}

		if name == "" && processName == "" {
			continue
		}

		out = append(out, model.SectionExcelRow{
			DepartmentID: 1,
			Name:         name,
			Color:        color,
			ProcessName:  processName,
		})
	}

	return out, nil
}

func parseSectionCell(cell string) (string, string) {
	cell = strings.TrimSpace(cell)
	if cell == "" {
		return "", ""
	}

	open := strings.LastIndex(cell, "(")
	close := strings.LastIndex(cell, ")")
	if open >= 0 && close > open {
		name := strings.TrimSpace(cell[:open])
		color := strings.TrimSpace(cell[open+1 : close])
		return name, color
	}
	return cell, ""
}

func nullableString(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
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
