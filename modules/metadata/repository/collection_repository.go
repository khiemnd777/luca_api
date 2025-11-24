package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/khiemnd777/andy_api/modules/metadata/model"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/lib/pq"
)

type CollectionWithFields struct {
	model.CollectionDTO
	Fields      []*model.FieldDTO `json:"fields,omitempty"`
	FieldsCount int               `json:"fields_count,omitempty"`
}

type CollectionRepository struct {
	DB *sql.DB
}

func NewCollectionRepository(db *sql.DB) *CollectionRepository {
	return &CollectionRepository{DB: db}
}

func colToDTO(c *model.Collection) *model.CollectionDTO {
	var sif *string
	if c.ShowIf != nil && c.ShowIf.Valid {
		sif = utils.CleanJSON(&c.ShowIf.String)
	}

	return &model.CollectionDTO{
		ID:     c.ID,
		Slug:   c.Slug,
		Name:   c.Name,
		ShowIf: sif,
	}
}

func evaluateShowIf(result *CollectionWithFields, entityData *map[string]any) *CollectionWithFields {
	if entityData != nil && result.ShowIf != nil && *result.ShowIf != "" {
		var cond customfields.ShowIfCondition
		if err := json.Unmarshal([]byte(*result.ShowIf), &cond); err == nil {
			ok := customfields.EvaluateShowIf(&cond, *entityData)
			if !ok {
				result.Fields = nil
				return result
			}
		}
	}

	return result
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
			SELECT id, slug, name, show_if
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
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.ShowIf); err != nil {
			return nil, 0, err
		}
		coldto := colToDTO(&c)

		list = append(list, CollectionWithFields{CollectionDTO: *coldto})
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

func (r *CollectionRepository) GetBySlug(ctx context.Context, slug string, withFields, table, form, showHidden bool, entityData *map[string]any) (*CollectionWithFields, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, slug, name, show_if FROM collections WHERE slug = $1
	`, slug)

	var c model.Collection

	if err := row.Scan(&c.ID, &c.Slug, &c.Name, &c.ShowIf); err != nil {
		return nil, err
	}

	coldto := colToDTO(&c)

	result := &CollectionWithFields{CollectionDTO: *coldto}
	if withFields {
		fields, _ := r.GetFieldsByCollectionID(ctx, c.ID, table, form, showHidden)
		result.Fields = fields
	}

	result = evaluateShowIf(result, entityData)

	return result, nil
}

func (r *CollectionRepository) GetByID(ctx context.Context, id int, withFields, table, form, showHidden bool, entityData *map[string]any) (*CollectionWithFields, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, slug, name, show_if FROM collections WHERE id = $1
	`, id)
	var c model.Collection
	if err := row.Scan(&c.ID, &c.Slug, &c.Name, &c.ShowIf); err != nil {
		return nil,
			err
	}

	coldto := colToDTO(&c)

	result := &CollectionWithFields{CollectionDTO: *coldto}
	if withFields {
		fields, _ := r.GetFieldsByCollectionID(ctx, id, table, form, showHidden)
		result.Fields = fields
	}

	result = evaluateShowIf(result, entityData)

	return result, nil
}

func (r *CollectionRepository) Create(ctx context.Context, slug, name string, showIf *sql.NullString) (*model.CollectionDTO, error) {
	row := r.DB.QueryRowContext(ctx, `
		INSERT INTO collections (slug, name, show_if) VALUES ($1, $2, $3) RETURNING id, slug, name, show_if
	`, slug, name, showIf)
	var c model.Collection
	if err := row.Scan(&c.ID, &c.Slug, &c.Name, &c.ShowIf); err != nil {
		return nil, err
	}
	dto := colToDTO(&c)
	return dto, nil
}

func (r *CollectionRepository) Update(
	ctx context.Context,
	id int,
	slug, name *string,
	showIf *sql.NullString,
) (*model.CollectionDTO, error) {

	setParts := []string{}
	args := []any{}

	// slug
	if slug != nil {
		setParts = append(setParts, fmt.Sprintf("slug=$%d", len(args)+1))
		args = append(args, *slug)
	}

	// name
	if name != nil {
		setParts = append(setParts, fmt.Sprintf("name=$%d", len(args)+1))
		args = append(args, *name)
	}

	// show_if
	if showIf != nil {
		setParts = append(setParts, fmt.Sprintf("show_if=$%d", len(args)+1))
		args = append(args, *showIf)
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("nothing to update")
	}

	// WHERE id = ...
	args = append(args, id)
	wherePos := len(args)

	query := fmt.Sprintf(`
		UPDATE collections
		SET %s
		WHERE id=$%d
		RETURNING id, slug, name, show_if
	`, strings.Join(setParts, ", "), wherePos)

	row := r.DB.QueryRowContext(ctx, query, args...)

	var c model.Collection
	if err := row.Scan(&c.ID, &c.Slug, &c.Name, &c.ShowIf); err != nil {
		return nil, err
	}

	return colToDTO(&c), nil
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
