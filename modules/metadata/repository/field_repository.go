package repository

import (
	"context"
	"database/sql"

	"github.com/khiemnd777/andy_api/modules/metadata/model"
)

type FieldRepository struct{ DB *sql.DB }

func NewFieldRepository(db *sql.DB) *FieldRepository { return &FieldRepository{DB: db} }

func (r *FieldRepository) ListByCollectionID(ctx context.Context, collectionID int) ([]model.Field, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, collection_id, name, label, type, required, "unique",
		       default_value, options, order_index, visibility, relation
		FROM fields
		WHERE collection_id=$1
		ORDER BY order_index ASC, id ASC
	`, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Field
	for rows.Next() {
		var f model.Field
		if err := rows.Scan(
			&f.ID, &f.CollectionID, &f.Name, &f.Label, &f.Type, &f.Required, &f.Unique,
			&f.DefaultValue, &f.Options, &f.OrderIndex, &f.Visibility, &f.Relation,
		); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *FieldRepository) Get(ctx context.Context, id int) (*model.Field, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, collection_id, name, label, type, required, "unique",
		       default_value, options, order_index, visibility, relation
		FROM fields WHERE id=$1
	`, id)
	var f model.Field
	if err := row.Scan(
		&f.ID, &f.CollectionID, &f.Name, &f.Label, &f.Type, &f.Required, &f.Unique,
		&f.DefaultValue, &f.Options, &f.OrderIndex, &f.Visibility, &f.Relation,
	); err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FieldRepository) Create(ctx context.Context, f *model.Field) (*model.Field, error) {
	row := r.DB.QueryRowContext(ctx, `
		INSERT INTO fields (collection_id, name, label, type, required, "unique",
		                    default_value, options, order_index, visibility, relation)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id
	`, f.CollectionID, f.Name, f.Label, f.Type, f.Required, f.Unique, f.DefaultValue, f.Options, f.OrderIndex, f.Visibility, f.Relation)
	if err := row.Scan(&f.ID); err != nil {
		return nil, err
	}
	return f, nil
}

func (r *FieldRepository) Update(ctx context.Context, f *model.Field) (*model.Field, error) {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE fields
		SET name=$1, label=$2, type=$3, required=$4, "unique"=$5,
		    default_value=$6, options=$7, order_index=$8, visibility=$9, relation=$10
		WHERE id=$11
	`, f.Name, f.Label, f.Type, f.Required, f.Unique, f.DefaultValue, f.Options, f.OrderIndex, f.Visibility, f.Relation, f.ID)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r *FieldRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM fields WHERE id=$1`, id)
	return err
}
