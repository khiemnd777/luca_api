package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
)

type ProcessImportRepository interface {
	GetOrCreate(ctx context.Context, deptID int, name string) (id int, created bool, err error)
}

type processImportRepo struct {
	db *sql.DB
}

func NewProcessImportRepository(db *sql.DB) ProcessImportRepository {
	return &processImportRepo{db: db}
}

type sqlRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (r *processImportRepo) runner(ctx context.Context) sqlRunner {
	if tx := txFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}

func (r *processImportRepo) GetOrCreate(ctx context.Context, deptID int, name string) (int, bool, error) {
	id, err := r.selectByName(ctx, deptID, name)
	if err == nil && id > 0 {
		return id, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	query := `
		INSERT INTO processes (department_id, name, active, custom_fields, created_at, updated_at)
		VALUES ($1, $2, TRUE, '{}'::jsonb, NOW(), NOW())
		RETURNING id
	`

	runner := r.runner(ctx)
	if err := runner.QueryRowContext(ctx, query, deptID, name).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			id, selErr := r.selectByName(ctx, deptID, name)
			if selErr != nil {
				return 0, false, selErr
			}
			return id, false, nil
		}
		return 0, false, err
	}

	return id, true, nil
}

func (r *processImportRepo) selectByName(ctx context.Context, deptID int, name string) (int, error) {
	query := `
		SELECT id
		FROM processes
		WHERE department_id = $1 AND name = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	var id int
	runner := r.runner(ctx)
	return id, runner.QueryRowContext(ctx, query, deptID, name).Scan(&id)
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
