package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
)

type TechniqueImportRepository interface {
	GetCategoryByName(ctx context.Context, deptID int, name string) (id int, categoryName string, err error)
	GetOrCreateTechnique(ctx context.Context, deptID, categoryID int, categoryName, name string) (id int, created bool, err error)
}

type techniqueImportRepo struct {
	db *sql.DB
}

func NewTechniqueImportRepository(db *sql.DB) TechniqueImportRepository {
	return &techniqueImportRepo{db: db}
}

type sqlRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (r *techniqueImportRepo) runner(ctx context.Context) sqlRunner {
	if tx := txFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}

func (r *techniqueImportRepo) GetCategoryByName(ctx context.Context, deptID int, name string) (int, string, error) {
	query := `
		SELECT id, name
		FROM categories
		WHERE name = $1
			AND deleted_at IS NULL
			AND (department_id = $2 OR department_id IS NULL)
		ORDER BY (department_id = $2) DESC, department_id NULLS LAST
		LIMIT 1
	`

	var id int
	var categoryName string
	runner := r.runner(ctx)
	return id, categoryName, runner.QueryRowContext(ctx, query, name, deptID).Scan(&id, &categoryName)
}

func (r *techniqueImportRepo) GetOrCreateTechnique(ctx context.Context, deptID, categoryID int, categoryName, name string) (int, bool, error) {
	id, err := r.selectByCategoryAndName(ctx, deptID, categoryID, name)
	if err == nil && id > 0 {
		return id, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	query := `
		INSERT INTO techniques (department_id, category_id, category_name, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`

	runner := r.runner(ctx)
	if err := runner.QueryRowContext(ctx, query, deptID, categoryID, categoryName, name).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			id, selErr := r.selectByCategoryAndName(ctx, deptID, categoryID, name)
			if selErr != nil {
				return 0, false, selErr
			}
			return id, false, nil
		}
		return 0, false, err
	}

	return id, true, nil
}

func (r *techniqueImportRepo) selectByCategoryAndName(ctx context.Context, deptID, categoryID int, name string) (int, error) {
	query := `
		SELECT id
		FROM techniques
		WHERE department_id = $1 AND category_id = $2 AND name = $3 AND deleted_at IS NULL
		LIMIT 1
	`

	var id int
	runner := r.runner(ctx)
	return id, runner.QueryRowContext(ctx, query, deptID, categoryID, name).Scan(&id)
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
