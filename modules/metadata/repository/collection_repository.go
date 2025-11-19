package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/metadata/model"
	"github.com/lib/pq"
)

type CollectionWithFields struct {
	model.Collection
	Fields      []*model.FieldDTO `json:"fields,omitempty"`
	FieldsCount int               `json:"fields_count,omitempty"`
}

type CollectionRepository struct {
	DB *sql.DB
}

func NewCollectionRepository(db *sql.DB) *CollectionRepository {
	return &CollectionRepository{DB: db}
}

func (r *CollectionRepository) List(ctx context.Context, query string, limit, offset int, withFields, table, form bool) ([]CollectionWithFields, int, error) {
	list := []CollectionWithFields{}
	var args []any
	where := ""
	if query != "" {
		where = "WHERE slug ILIKE $1 OR name ILIKE $1"
		args = append(args, "%"+query+"%")
	}

	rows, err := r.DB.QueryContext(ctx,
		fmt.Sprintf(`
			SELECT id, slug, name
			FROM collections
			%s
			ORDER BY slug ASC
			LIMIT %d OFFSET %d
		`, where, limit, offset), args...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var c model.Collection
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name); err != nil {
			return nil, 0, err
		}
		list = append(list, CollectionWithFields{Collection: c})
	}

	// count
	var total int
	if err := r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM collections").Scan(&total); err != nil {
		total = len(list)
	}

	ids := make([]int, len(list))
	for i := range list {
		ids[i] = list[i].ID
	}

	counts, err := r.GetFieldCountsBatch(ctx, ids, table, form)
	if err != nil {
		return nil, 0, err
	}

	for i := range list {
		list[i].FieldsCount = counts[list[i].ID]

		if withFields {
			fields, err := r.GetFieldsByCollectionID(ctx, list[i].ID, table, form, true)
			if err != nil {
				return nil, 0, err
			}
			list[i].Fields = fields
		}
	}

	return list, total, nil
}

func (r *CollectionRepository) GetBySlug(ctx context.Context, slug string, withFields, table, form, showHidden bool) (*CollectionWithFields, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, slug, name FROM collections WHERE slug = $1
	`, slug)
	var c model.Collection
	if err := row.Scan(&c.ID, &c.Slug, &c.Name); err != nil {
		return nil, err
	}

	result := &CollectionWithFields{Collection: c}
	if withFields {
		fields, _ := r.GetFieldsByCollectionID(ctx, c.ID, table, form, showHidden)
		result.Fields = fields
	}
	return result, nil
}

func (r *CollectionRepository) GetByID(ctx context.Context, id int, withFields, table, form, showHidden bool) (*CollectionWithFields, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, slug, name FROM collections WHERE id = $1
	`, id)
	var c model.Collection
	if err := row.Scan(&c.ID, &c.Slug, &c.Name); err != nil {
		return nil, err
	}
	result := &CollectionWithFields{Collection: c}
	if withFields {
		fields, _ := r.GetFieldsByCollectionID(ctx, id, table, form, showHidden)
		result.Fields = fields
	}
	return result, nil
}

func (r *CollectionRepository) Create(ctx context.Context, slug, name string) (*model.Collection, error) {
	row := r.DB.QueryRowContext(ctx, `
		INSERT INTO collections (slug, name) VALUES ($1, $2) RETURNING id, slug, name
	`, slug, name)
	var c model.Collection
	if err := row.Scan(&c.ID, &c.Slug, &c.Name); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CollectionRepository) Update(ctx context.Context, id int, slug, name *string) (*model.Collection, error) {
	q := "UPDATE collections SET "
	args := []any{}
	if slug != nil {
		q += "slug=$1"
		args = append(args, *slug)
	}
	if name != nil {
		if len(args) > 0 {
			q += ", "
		}
		args = append(args, *name)
		q += fmt.Sprintf("name=$%d", len(args))
	}
	args = append(args, id)
	q += fmt.Sprintf(" WHERE id=$%d RETURNING id, slug, name", len(args))

	row := r.DB.QueryRowContext(ctx, q, args...)
	var c model.Collection
	if err := row.Scan(&c.ID, &c.Slug, &c.Name); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CollectionRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM collections WHERE id=$1`, id)
	return err
}

func (r *CollectionRepository) SlugExists(ctx context.Context, slug string, excludeID *int) (bool, error) {
	q := "SELECT COUNT(*) FROM collections WHERE slug=$1"
	args := []any{slug}
	if excludeID != nil {
		q += fmt.Sprintf(" AND id <> $%d", len(args)+1)
		args = append(args, *excludeID)
	}
	var cnt int
	if err := r.DB.QueryRowContext(ctx, q, args...).Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (r *CollectionRepository) GetFieldsByCollectionID(ctx context.Context, collectionID int, table, form, showHidden bool) ([]*model.FieldDTO, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT 
			id, 
			collection_id, 
			name, 
			label, 
			type, 
			required, 
			"unique", 
			"table", 
			form, 
			search,
			default_value, 
			options, 
			order_index, 
			visibility, 
			relation
		FROM fields
		WHERE collection_id = $1 
			AND ($2::bool IS FALSE OR "table" = TRUE)
			AND ($3::bool IS FALSE OR form = TRUE)
			AND ($4::bool IS TRUE OR visibility IS DISTINCT FROM 'hidden')
		ORDER BY order_index ASC
	`, collectionID, table, form, showHidden)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.Field{}
	for rows.Next() {
		var f model.Field
		if err := rows.Scan(
			&f.ID,
			&f.CollectionID,
			&f.Name,
			&f.Label,
			&f.Type,
			&f.Required,
			&f.Unique,
			&f.Table,
			&f.Form,
			&f.Search,
			&f.DefaultValue,
			&f.Options,
			&f.OrderIndex,
			&f.Visibility,
			&f.Relation,
		); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return fieldsToDTOs(list), nil
}

func (r *CollectionRepository) GetFieldCountsBatch(ctx context.Context, collectionIDs []int, table, form bool) (map[int]int, error) {
	rows, err := r.DB.QueryContext(ctx, `
        SELECT collection_id, COUNT(*)
        FROM fields
        WHERE collection_id = ANY($1)
          AND ($2::bool IS FALSE OR "table" = TRUE)
          AND ($3::bool IS FALSE OR form = TRUE)
        GROUP BY collection_id
    `, pq.Array(collectionIDs), table, form)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var cid, cnt int
		if err := rows.Scan(&cid, &cnt); err != nil {
			return nil, err
		}
		result[cid] = cnt
	}
	return result, nil
}
