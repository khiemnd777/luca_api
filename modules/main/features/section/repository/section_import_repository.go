package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
)

type SectionImportRepository interface {
	GetOrCreateSection(ctx context.Context, deptID int, name string, color *string) (id int, created bool, err error)
	GetProcessByName(ctx context.Context, deptID int, name string) (id int, nameOut string, err error)
	UpsertSectionProcess(ctx context.Context, sectionID int, sectionName string, processID int, processName string, color *string, displayOrder int) (bool, error)
	UpdateProcessSectionCache(ctx context.Context, deptID int, processID int, sectionID int, sectionName string, color *string) error
	UpdateSectionProcessNames(ctx context.Context, sectionID int) error
}

type sectionImportRepo struct {
	db *sql.DB
}

func NewSectionImportRepository(db *sql.DB) SectionImportRepository {
	return &sectionImportRepo{db: db}
}

type sqlRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func (r *sectionImportRepo) runner(ctx context.Context) sqlRunner {
	if tx := txFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}

func (r *sectionImportRepo) GetOrCreateSection(ctx context.Context, deptID int, name string, color *string) (int, bool, error) {
	id, err := r.selectSectionByName(ctx, deptID, name)
	if err == nil && id > 0 {
		return id, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	query := `
		INSERT INTO sections (department_id, name, color, active, custom_fields, created_at, updated_at)
		VALUES ($1, $2, $3, TRUE, '{}'::jsonb, NOW(), NOW())
		RETURNING id
	`

	runner := r.runner(ctx)
	if err := runner.QueryRowContext(ctx, query, deptID, name, color).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			id, selErr := r.selectSectionByName(ctx, deptID, name)
			if selErr != nil {
				return 0, false, selErr
			}
			return id, false, nil
		}
		return 0, false, err
	}

	return id, true, nil
}

func (r *sectionImportRepo) selectSectionByName(ctx context.Context, deptID int, name string) (int, error) {
	query := `
		SELECT id
		FROM sections
		WHERE department_id = $1 AND name = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	var id int
	runner := r.runner(ctx)
	return id, runner.QueryRowContext(ctx, query, deptID, name).Scan(&id)
}

func (r *sectionImportRepo) GetProcessByName(ctx context.Context, deptID int, name string) (int, string, error) {
	query := `
		SELECT id, name
		FROM processes
		WHERE department_id = $1 AND name = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	var id int
	var outName string
	runner := r.runner(ctx)
	return id, outName, runner.QueryRowContext(ctx, query, deptID, name).Scan(&id, &outName)
}

func (r *sectionImportRepo) UpsertSectionProcess(ctx context.Context, sectionID int, sectionName string, processID int, processName string, color *string, displayOrder int) (bool, error) {
	query := `
		INSERT INTO section_processes (section_id, process_id, section_name, process_name, color, display_order, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (section_id, process_id)
		DO UPDATE SET
			section_name = EXCLUDED.section_name,
			process_name = EXCLUDED.process_name,
			color = EXCLUDED.color,
			display_order = EXCLUDED.display_order
		RETURNING (xmax = 0)
	`

	var inserted bool
	runner := r.runner(ctx)
	if err := runner.QueryRowContext(ctx, query, sectionID, processID, sectionName, processName, color, displayOrder).Scan(&inserted); err != nil {
		return false, err
	}
	return inserted, nil
}

func (r *sectionImportRepo) UpdateProcessSectionCache(ctx context.Context, deptID int, processID int, sectionID int, sectionName string, color *string) error {
	query := `
		UPDATE processes
		SET section_id = $1,
			section_name = $2,
			color = $3,
			updated_at = NOW()
		WHERE id = $4 AND department_id = $5
	`

	_, err := r.runner(ctx).ExecContext(ctx, query, sectionID, sectionName, color, processID, deptID)
	return err
}

func (r *sectionImportRepo) UpdateSectionProcessNames(ctx context.Context, sectionID int) error {
	query := `
		UPDATE sections
		SET process_names = (
			SELECT string_agg(sp.process_name, ', ' ORDER BY sp.display_order, sp.process_name)
			FROM section_processes sp
			WHERE sp.section_id = sections.id
		)
		WHERE id = $1
	`

	_, err := r.runner(ctx).ExecContext(ctx, query, sectionID)
	return err
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key value") || strings.Contains(msg, "unique constraint")
}

func txFromContext(ctx context.Context) *sql.Tx {
	if ctx == nil {
		return nil
	}
	if tx, ok := ctx.Value(txContextKey{}).(*sql.Tx); ok {
		return tx
	}
	return nil
}

type txContextKey struct{}

func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, txContextKey{}, tx)
}
