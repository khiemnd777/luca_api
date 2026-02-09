package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
)

type RestorationTypeImportRepository interface {
	GetCategoryIDByName(ctx context.Context, deptID int, name string) (int, error)
	GetOrCreateRestorationType(ctx context.Context, deptID int, categoryID int, categoryName string, name string) (id int, created bool, err error)
}

type restorationTypeImportRepo struct {
	db *sql.DB
}

func NewRestorationTypeImportRepository(db *sql.DB) RestorationTypeImportRepository {
	return &restorationTypeImportRepo{db: db}
}

type sqlRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (r *restorationTypeImportRepo) runner(ctx context.Context) sqlRunner {
	if tx := txFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}

func (r *restorationTypeImportRepo) GetCategoryIDByName(ctx context.Context, deptID int, name string) (int, error) {
	query := `
		SELECT id
		FROM categories
		WHERE department_id = $1 AND name = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	var id int
	runner := r.runner(ctx)
	return id, runner.QueryRowContext(ctx, query, deptID, name).Scan(&id)
}

func (r *restorationTypeImportRepo) GetOrCreateRestorationType(ctx context.Context, deptID int, categoryID int, categoryName string, name string) (int, bool, error) {
	id, err := r.selectByCategoryAndName(ctx, deptID, categoryID, name)
	if err == nil && id > 0 {
		return id, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	query := `
		INSERT INTO restoration_types (category_id, category_name, name, department_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`

	runner := r.runner(ctx)
	if err := runner.QueryRowContext(ctx, query, categoryID, categoryName, name, deptID).Scan(&id); err != nil {
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

func (r *restorationTypeImportRepo) selectByCategoryAndName(ctx context.Context, deptID int, categoryID int, name string) (int, error) {
	query := `
		SELECT id
		FROM restoration_types
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
