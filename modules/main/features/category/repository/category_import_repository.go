package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	collectionutils "github.com/khiemnd777/andy_api/shared/metadata/collection"
)

type CategoryImportRepository interface {
	GetOrCreateLV1(ctx context.Context, deptID int, name string) (id int, created bool, err error)
	GetOrCreateLV2(ctx context.Context, deptID int, lv1ID int, lv1Name, name string) (id int, created bool, err error)
	GetOrCreateLV3(ctx context.Context, deptID int, lv1ID, lv2ID int, lv1Name, lv2Name, name string) (id int, created bool, err error)
	GetTreeNode(ctx context.Context, deptID int, id int) (*collectionutils.TreeNode, error)
	GetCollectionID(ctx context.Context, deptID int, id int) (*int, error)
	UpsertFields(ctx context.Context, collectionID int, fields []CategoryFieldSpec) (int, error)
}

type categoryImportRepo struct {
	db *generated.Client
}

func NewCategoryImportRepository(db *generated.Client) CategoryImportRepository {
	return &categoryImportRepo{db: db}
}

type sqlRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
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
	if err := queryRow(ctx, runner, query, []any{&id}, name, deptID); err != nil {
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
	if err := queryRow(ctx, runner, query, []any{&id}, name, lv1ID, lv1ID, lv1Name, deptID); err != nil {
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
	if err := queryRow(ctx, runner, query, []any{&id}, name, lv2ID, lv1ID, lv1Name, lv2ID, lv2Name, deptID); err != nil {
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
	return id, queryRow(ctx, runner, query, []any{&id}, deptID, name)
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
	return id, queryRow(ctx, runner, query, []any{&id}, deptID, parentID, name)
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
	return id, queryRow(ctx, runner, query, []any{&id}, deptID, parentID, name)
}

func (r *categoryImportRepo) GetTreeNode(ctx context.Context, deptID int, id int) (*collectionutils.TreeNode, error) {
	query := `
		SELECT id, parent_id, name, collection_id
		FROM categories
		WHERE department_id = $1::INT AND id = $2 AND deleted_at IS NULL
		LIMIT 1
	`

	var node collectionutils.TreeNode
	var parentID sql.NullInt64
	var name sql.NullString
	var collectionID sql.NullInt64

	runner := r.runner(ctx)
	if err := queryRow(ctx, runner, query, []any{&node.ID, &parentID, &name, &collectionID}, deptID, id); err != nil {
		return nil, err
	}

	if parentID.Valid {
		v := int(parentID.Int64)
		node.ParentID = &v
	}
	if name.Valid {
		v := name.String
		node.Name = &v
	}
	if collectionID.Valid {
		v := int(collectionID.Int64)
		node.CollectionID = &v
	}
	return &node, nil
}

func (r *categoryImportRepo) GetCollectionID(ctx context.Context, deptID int, id int) (*int, error) {
	query := `
		SELECT collection_id
		FROM categories
		WHERE department_id = $1::INT AND id = $2 AND deleted_at IS NULL
		LIMIT 1
	`
	var collectionID sql.NullInt64
	runner := r.runner(ctx)
	if err := queryRow(ctx, runner, query, []any{&collectionID}, deptID, id); err != nil {
		return nil, err
	}
	if !collectionID.Valid {
		return nil, nil
	}
	v := int(collectionID.Int64)
	return &v, nil
}

type CategoryFieldSpec struct {
	Name         string
	Label        string
	Type         string
	Required     bool
	Unique       bool
	Tag          *string
	Table        bool
	Form         bool
	Search       bool
	DefaultValue *string
	Options      *string
	OrderIndex   int
	Visibility   string
	Relation     *string
}

func (r *categoryImportRepo) UpsertFields(ctx context.Context, collectionID int, fields []CategoryFieldSpec) (int, error) {
	if len(fields) == 0 {
		return 0, nil
	}

	runner := r.runner(ctx)
	changed := 0

	for _, f := range fields {
		var existsID int
		if err := queryRow(ctx, runner, `
			SELECT id FROM fields WHERE collection_id = $1 AND name = $2 LIMIT 1
		`, []any{&existsID}, collectionID, f.Name); err == nil {
			_, err := runner.ExecContext(ctx, `
				UPDATE fields
				SET name=$1,
				    label=$2,
				    type=$3,
				    required=$4,
				    "unique"=$5,
				    tag=$6,
				    "table"=$7,
				    form=$8,
				    search=$9,
				    default_value=$10,
				    options=$11,
				    order_index=$12,
				    visibility=$13,
				    relation=$14
				WHERE id=$15
			`,
				f.Name,
				f.Label,
				f.Type,
				f.Required,
				f.Unique,
				f.Tag,
				f.Table,
				f.Form,
				f.Search,
				f.DefaultValue,
				f.Options,
				f.OrderIndex,
				f.Visibility,
				f.Relation,
				existsID,
			)
			if err != nil {
				return changed, err
			}
			changed++
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			return changed, err
		}

		_, err := runner.ExecContext(ctx, `
			INSERT INTO fields (
				collection_id,
				name,
				label,
				type,
				required,
				"unique",
				tag,
				"table",
				form,
				search,
				default_value,
				options,
				order_index,
				visibility,
				relation
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		`,
			collectionID,
			f.Name,
			f.Label,
			f.Type,
			f.Required,
			f.Unique,
			f.Tag,
			f.Table,
			f.Form,
			f.Search,
			f.DefaultValue,
			f.Options,
			f.OrderIndex,
			f.Visibility,
			f.Relation,
		)
		if err != nil {
			return changed, err
		}
		changed++
	}

	return changed, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key value") || strings.Contains(msg, "unique constraint")
}

func txFromContext(ctx context.Context) *generated.Tx {
	if ctx == nil {
		return nil
	}
	if tx, ok := ctx.Value(txContextKey{}).(*generated.Tx); ok {
		return tx
	}
	return nil
}

type txContextKey struct{}

func WithTx(ctx context.Context, tx *generated.Tx) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, txContextKey{}, tx)
}

func queryRow(ctx context.Context, runner sqlRunner, query string, dests []any, args ...any) error {
	rows, err := runner.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	return rows.Scan(dests...)
}
