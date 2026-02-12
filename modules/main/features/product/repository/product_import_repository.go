package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/lib/pq"
)

type ProductImportRepository interface {
	FindProductByCode(ctx context.Context, deptID int, code string) (*ProductImportProductRef, error)
	ResolveCategoryBranch(ctx context.Context, deptID int, lv1, lv2, lv3 string) (id int, name string, err error)
	ResolveCategoryLV1(ctx context.Context, deptID int, lv1 string) (id int, name string, err error)
	GetOrCreateBrandName(ctx context.Context, deptID int, categoryID int, categoryName, name string) (id int, created bool, err error)
	GetOrCreateRawMaterial(ctx context.Context, deptID int, categoryID int, categoryName, name string) (id int, created bool, err error)
	GetOrCreateTechnique(ctx context.Context, deptID int, categoryID int, categoryName, name string) (id int, created bool, err error)
	GetOrCreateRestorationType(ctx context.Context, deptID int, categoryID int, categoryName, name string) (id int, created bool, err error)
	GetOrCreateProcess(ctx context.Context, deptID int, name string) (id int, created bool, err error)
}

type ProductImportProductRef struct {
	ID         int
	TemplateID *int
}

type productImportRepo struct {
	db *sql.DB
}

func NewProductImportRepository(db *sql.DB) ProductImportRepository {
	return &productImportRepo{db: db}
}

func (r *productImportRepo) FindProductByCode(ctx context.Context, deptID int, code string) (*ProductImportProductRef, error) {
	query := `
		SELECT id, template_id
		FROM products
		WHERE department_id = $1
			AND code_norm = lower(unaccent_immutable($2))
			AND deleted_at IS NULL
		LIMIT 1
	`

	var out ProductImportProductRef
	var templateID sql.NullInt64
	if err := r.db.QueryRowContext(ctx, query, deptID, code).Scan(&out.ID, &templateID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if templateID.Valid {
		v := int(templateID.Int64)
		out.TemplateID = &v
	}
	return &out, nil
}

func (r *productImportRepo) ResolveCategoryBranch(ctx context.Context, deptID int, lv1, lv2, lv3 string) (int, string, error) {
	lv1ID, lv1Name, err := r.selectCategory(ctx, deptID, 1, nil, lv1)
	if err != nil {
		return 0, "", err
	}
	targetID := lv1ID
	pathParts := []string{lv1Name}

	if lv2 != "" {
		lv2ID, lv2Name, err := r.selectCategory(ctx, deptID, 2, &lv1ID, lv2)
		if err != nil {
			return 0, "", err
		}
		targetID = lv2ID
		pathParts = append(pathParts, lv2Name)

		if lv3 != "" {
			lv3ID, lv3Name, err := r.selectCategory(ctx, deptID, 3, &lv2ID, lv3)
			if err != nil {
				return 0, "", err
			}
			targetID = lv3ID
			pathParts = append(pathParts, lv3Name)
		}
	}

	return targetID, strings.Join(pathParts, " > "), nil
}

func (r *productImportRepo) ResolveCategoryLV1(ctx context.Context, deptID int, lv1 string) (int, string, error) {
	return r.selectCategory(ctx, deptID, 1, nil, lv1)
}

func (r *productImportRepo) selectCategory(
	ctx context.Context,
	deptID int,
	level int,
	parentID *int,
	name string,
) (int, string, error) {

	name = strings.Join(strings.Fields(name), " ")
	norm := utils.NormalizeSearchKeyword(name)

	var (
		id          int
		displayName string
	)

	// Try strict match first
	id, displayName, err := r.queryCategory(ctx, deptID, level, parentID, norm, false)
	if err == nil {
		return id, displayName, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, "", err
	}

	// Fallback: ignore whitespace
	return r.queryCategory(ctx, deptID, level, parentID, norm, true)
}

func (r *productImportRepo) queryCategory(
	ctx context.Context,
	deptID int,
	level int,
	parentID *int,
	norm string,
	ignoreWhitespace bool,
) (int, string, error) {

	var (
		query string
		args  []any
	)

	query = `
		SELECT id, name
		FROM categories
		WHERE level = $1
			AND department_id = $2
			AND deleted_at IS NULL
	`
	args = []any{level, deptID}

	if ignoreWhitespace {
		query += `
			AND regexp_replace(name_norm, '\s+', '', 'g')
			    = regexp_replace($3, '\s+', '', 'g')
		`
		args = append(args, norm)
	} else {
		query += ` AND name_norm = $3 `
		args = append(args, norm)
	}

	if parentID == nil {
		query += ` AND parent_id IS NULL `
	} else {
		query += ` AND parent_id = $4 `
		args = append(args, *parentID)
	}

	query += ` LIMIT 1 `

	var id int
	var displayName string

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&id, &displayName)
	return id, displayName, err
}

func (r *productImportRepo) GetOrCreateBrandName(ctx context.Context, deptID int, categoryID int, categoryName, name string) (int, bool, error) {
	id, err := r.selectRefByCategoryAndName(ctx, "brand_names", deptID, categoryID, name)
	if err == nil && id > 0 {
		return id, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	query := `
		INSERT INTO brand_names (department_id, category_id, category_name, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`
	if err := r.db.QueryRowContext(ctx, query, deptID, categoryID, categoryName, name).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			id, selErr := r.selectRefByCategoryAndName(ctx, "brand_names", deptID, categoryID, name)
			if selErr != nil {
				return 0, false, selErr
			}
			return id, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}

func (r *productImportRepo) GetOrCreateRawMaterial(ctx context.Context, deptID int, categoryID int, categoryName, name string) (int, bool, error) {
	id, err := r.selectRefByCategoryAndName(ctx, "raw_materials", deptID, categoryID, name)
	if err == nil && id > 0 {
		return id, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	query := `
		INSERT INTO raw_materials (department_id, category_id, category_name, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`
	if err := r.db.QueryRowContext(ctx, query, deptID, categoryID, categoryName, name).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			id, selErr := r.selectRefByCategoryAndName(ctx, "raw_materials", deptID, categoryID, name)
			if selErr != nil {
				return 0, false, selErr
			}
			return id, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}

func (r *productImportRepo) GetOrCreateTechnique(ctx context.Context, deptID int, categoryID int, categoryName, name string) (int, bool, error) {
	id, err := r.selectRefByCategoryAndName(ctx, "techniques", deptID, categoryID, name)
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
	if err := r.db.QueryRowContext(ctx, query, deptID, categoryID, categoryName, name).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			id, selErr := r.selectRefByCategoryAndName(ctx, "techniques", deptID, categoryID, name)
			if selErr != nil {
				return 0, false, selErr
			}
			return id, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}

func (r *productImportRepo) GetOrCreateRestorationType(ctx context.Context, deptID int, categoryID int, categoryName, name string) (int, bool, error) {
	id, err := r.selectRefByCategoryAndName(ctx, "restoration_types", deptID, categoryID, name)
	if err == nil && id > 0 {
		return id, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	query := `
		INSERT INTO restoration_types (department_id, category_id, category_name, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`
	if err := r.db.QueryRowContext(ctx, query, deptID, categoryID, categoryName, name).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			id, selErr := r.selectRefByCategoryAndName(ctx, "restoration_types", deptID, categoryID, name)
			if selErr != nil {
				return 0, false, selErr
			}
			return id, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}

func (r *productImportRepo) GetOrCreateProcess(ctx context.Context, deptID int, name string) (int, bool, error) {
	id, err := r.selectProcessByName(ctx, deptID, name)
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
	if err := r.db.QueryRowContext(ctx, query, deptID, name).Scan(&id); err != nil {
		if isUniqueViolation(err) {
			id, selErr := r.selectProcessByName(ctx, deptID, name)
			if selErr != nil {
				return 0, false, selErr
			}
			return id, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}

func (r *productImportRepo) selectProcessByName(ctx context.Context, deptID int, name string) (int, error) {
	query := `
		SELECT id
		FROM processes
		WHERE department_id = $1
			AND name_norm = lower(unaccent_immutable($2))
			AND deleted_at IS NULL
		LIMIT 1
	`
	var id int
	return id, r.db.QueryRowContext(ctx, query, deptID, name).Scan(&id)
}

func (r *productImportRepo) selectRefByCategoryAndName(ctx context.Context, table string, deptID int, categoryID int, name string) (int, error) {
	query := `
		SELECT id
		FROM ` + table + `
		WHERE department_id = $1
			AND category_id = $2
			AND name_norm = lower(unaccent_immutable($3))
			AND deleted_at IS NULL
		LIMIT 1
	`

	var id int
	return id, r.db.QueryRowContext(ctx, query, deptID, categoryID, name).Scan(&id)
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key value") || strings.Contains(msg, "unique constraint")
}
