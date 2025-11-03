package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type Collection struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Field struct {
	ID           int            `json:"id"`
	CollectionID int            `json:"collection_id"`
	Name         string         `json:"name"`
	Label        string         `json:"label"`
	Type         string         `json:"type"`
	Required     bool           `json:"required"`
	Unique       bool           `json:"unique"`
	DefaultValue sql.NullString `json:"default_value"`
	Options      sql.NullString `json:"options"`
	OrderIndex   int            `json:"order_index"`
	Visibility   string         `json:"visibility"`
	Relation     sql.NullString `json:"relation"`
}

type CollectionWithFields struct {
	Collection
	Fields []Field `json:"fields,omitempty"`
}

type CollectionRepository struct {
	DB *sql.DB
}

func NewCollectionRepository(db *sql.DB) *CollectionRepository {
	return &CollectionRepository{DB: db}
}

func (r *CollectionRepository) List(ctx context.Context, query string, limit, offset int, withFields bool) ([]CollectionWithFields, int, error) {
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
		var c Collection
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

	// preload fields nếu cần
	if withFields {
		for i := range list {
			f, _ := r.GetFieldsByCollectionID(ctx, list[i].ID)
			list[i].Fields = f
		}
	}
	return list, total, nil
}

func (r *CollectionRepository) GetBySlug(ctx context.Context, slug string, withFields bool) (*CollectionWithFields, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, slug, name FROM collections WHERE slug = $1
	`, slug)
	var c Collection
	if err := row.Scan(&c.ID, &c.Slug, &c.Name); err != nil {
		return nil, err
	}

	result := &CollectionWithFields{Collection: c}
	if withFields {
		fields, _ := r.GetFieldsByCollectionID(ctx, c.ID)
		result.Fields = fields
	}
	return result, nil
}

func (r *CollectionRepository) GetByID(ctx context.Context, id int, withFields bool) (*CollectionWithFields, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, slug, name FROM collections WHERE id = $1
	`, id)
	var c Collection
	if err := row.Scan(&c.ID, &c.Slug, &c.Name); err != nil {
		return nil, err
	}
	result := &CollectionWithFields{Collection: c}
	if withFields {
		fields, _ := r.GetFieldsByCollectionID(ctx, id)
		result.Fields = fields
	}
	return result, nil
}

func (r *CollectionRepository) Create(ctx context.Context, slug, name string) (*Collection, error) {
	row := r.DB.QueryRowContext(ctx, `
		INSERT INTO collections (slug, name) VALUES ($1, $2) RETURNING id, slug, name
	`, slug, name)
	var c Collection
	if err := row.Scan(&c.ID, &c.Slug, &c.Name); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CollectionRepository) Update(ctx context.Context, id int, slug, name *string) (*Collection, error) {
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
	var c Collection
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

func (r *CollectionRepository) GetFieldsByCollectionID(ctx context.Context, collectionID int) ([]Field, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, collection_id, name, label, type, required, "unique", default_value, options, order_index, visibility, relation
		FROM fields
		WHERE collection_id = $1
		ORDER BY order_index ASC
	`, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Field{}
	for rows.Next() {
		var f Field
		if err := rows.Scan(&f.ID, &f.CollectionID, &f.Name, &f.Label, &f.Type,
			&f.Required, &f.Unique, &f.DefaultValue, &f.Options, &f.OrderIndex, &f.Visibility, &f.Relation); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	return list, nil
}
