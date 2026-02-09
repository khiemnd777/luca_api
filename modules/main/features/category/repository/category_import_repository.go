package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
)

type CategoryImportRepository interface {
	GetOrCreateLV1(ctx context.Context, deptID int, name string) (id int, created bool, err error)
	GetOrCreateLV2(ctx context.Context, deptID int, lv1ID int, lv1Name, name string) (id int, created bool, err error)
	GetOrCreateLV3(ctx context.Context, deptID int, lv1ID, lv2ID int, lv1Name, lv2Name, name string) (id int, created bool, err error)
}

type categoryImportRepo struct {
	db *sql.DB
}

func NewCategoryImportRepository(db *sql.DB) CategoryImportRepository {
	return &categoryImportRepo{db: db}
}

type sqlRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (r *categoryImportRepo) runner(ctx context.Context) sqlRunner {
	if tx := txFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}

func (r *categoryImportRepo) GetOrCreateLV1(ctx context.Context, deptID int, name string) (int, bool, error) {
	id, err := r.selectLV1(ctx, deptID, name)
	if err == nil && id > 0 {
		return id, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	query := `
		INSERT INTO categories (name, level, active, custom_fields, department_id, created_at, updated_at)
		VALUES ($1, 1, TRUE, '{}'::jsonb, $2, NOW(), NOW())
		RETURNING id
	`

	runner := r.runner(ctx)
	if err := runner.QueryRowContext(ctx, query, name, deptID).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			id, selErr := r.selectLV1(ctx, deptID, name)
			if selErr != nil {
				return 0, false, selErr
			}
			return id, false, nil
		}
		return 0, false, err
	}

	return id, true, nil
}

func (r *categoryImportRepo) GetOrCreateLV2(ctx context.Context, deptID int, lv1ID int, lv1Name, name string) (int, bool, error) {
	id, err := r.selectLV2(ctx, deptID, lv1ID, name)
	if err == nil && id > 0 {
		return id, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	query := `
		INSERT INTO categories (
			name, level, parent_id,
			category_id_lv1, category_name_lv1,
			active, custom_fields, department_id, created_at, updated_at
		)
		VALUES ($1, 2, $2, $3, $4, TRUE, '{}'::jsonb, $5, NOW(), NOW())
		RETURNING id
	`

	runner := r.runner(ctx)
	if err := runner.QueryRowContext(ctx, query, name, lv1ID, lv1ID, lv1Name, deptID).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			id, selErr := r.selectLV2(ctx, deptID, lv1ID, name)
			if selErr != nil {
				return 0, false, selErr
			}
			return id, false, nil
		}
		return 0, false, err
	}

	return id, true, nil
}

func (r *categoryImportRepo) GetOrCreateLV3(ctx context.Context, deptID int, lv1ID, lv2ID int, lv1Name, lv2Name, name string) (int, bool, error) {
	id, err := r.selectLV3(ctx, deptID, lv2ID, name)
	if err == nil && id > 0 {
		return id, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	query := `
		INSERT INTO categories (
			name, level, parent_id,
			category_id_lv1, category_name_lv1,
			category_id_lv2, category_name_lv2,
			active, custom_fields, department_id, created_at, updated_at
		)
		VALUES ($1, 3, $2, $3, $4, $5, $6, TRUE, '{}'::jsonb, $7, NOW(), NOW())
		RETURNING id
	`

	runner := r.runner(ctx)
	if err := runner.QueryRowContext(ctx, query, name, lv2ID, lv1ID, lv1Name, lv2ID, lv2Name, deptID).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			id, selErr := r.selectLV3(ctx, deptID, lv2ID, name)
			if selErr != nil {
				return 0, false, selErr
			}
			return id, false, nil
		}
		return 0, false, err
	}

	return id, true, nil
}

func (r *categoryImportRepo) selectLV1(ctx context.Context, deptID int, name string) (int, error) {
	query := `
		SELECT id
		FROM categories
		WHERE department_id = $1::INT AND level = 1 AND name = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	var id int
	runner := r.runner(ctx)
	return id, runner.QueryRowContext(ctx, query, deptID, name).Scan(&id)
}

func (r *categoryImportRepo) selectLV2(ctx context.Context, deptID int, parentID int, name string) (int, error) {
	query := `
		SELECT id
		FROM categories
		WHERE department_id = $1::INT AND level = 2 AND parent_id = $2 AND name = $3 AND deleted_at IS NULL
		LIMIT 1
	`

	var id int
	runner := r.runner(ctx)
	return id, runner.QueryRowContext(ctx, query, deptID, parentID, name).Scan(&id)
}

func (r *categoryImportRepo) selectLV3(ctx context.Context, deptID int, parentID int, name string) (int, error) {
	query := `
		SELECT id
		FROM categories
		WHERE department_id = $1::INT AND level = 3 AND parent_id = $2 AND name = $3 AND deleted_at IS NULL
		LIMIT 1
	`

	var id int
	runner := r.runner(ctx)
	return id, runner.QueryRowContext(ctx, query, deptID, parentID, name).Scan(&id)
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
