package repository

import (
	"context"
	"database/sql"

	"github.com/khiemnd777/andy_api/modules/metadata/model"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type FieldRepository struct{ DB *sql.DB }

func NewFieldRepository(db *sql.DB) *FieldRepository { return &FieldRepository{DB: db} }

func fieldToDTO(f *model.Field) *model.FieldDTO {
	var dv *string
	if f.DefaultValue != nil && f.DefaultValue.Valid {
		dv = utils.CleanJSON(&f.DefaultValue.String)
	}

	var opt *string
	if f.Options != nil && f.Options.Valid {
		opt = utils.CleanJSON(&f.Options.String)
	}

	var rel *string
	if f.Relation != nil && f.Relation.Valid {
		rel = utils.CleanJSON(&f.Relation.String)
	}

	return &model.FieldDTO{
		ID:             f.ID,
		CollectionID:   f.CollectionID,
		CollectionSlug: f.CollectionSlug,
		Name:           f.Name,
		Label:          f.Label,
		Type:           f.Type,
		Required:       f.Required,
		Unique:         f.Unique,
		Tag:            f.Tag,
		Table:          f.Table,
		Form:           f.Form,
		Search:         f.Search,
		DefaultValue:   dv,
		Options:        opt,
		OrderIndex:     f.OrderIndex,
		Visibility:     f.Visibility,
		Relation:       rel,
	}
}

func fieldsToDTOs(list []model.Field) []*model.FieldDTO {
	out := make([]*model.FieldDTO, 0, len(list))
	for i := range list {
		dto := fieldToDTO(&list[i])
		out = append(out, dto)
	}
	return out
}

func (r *FieldRepository) ListByCollectionID(ctx context.Context, collectionID int) ([]*model.FieldDTO, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT 
			id, 
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
			&f.ID,
			&f.CollectionID,
			&f.Name,
			&f.Label,
			&f.Type,
			&f.Required,
			&f.Unique,
			&f.Tag,
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
		out = append(out, f)
	}

	return fieldsToDTOs(out), rows.Err()
}

func (r *FieldRepository) Get(ctx context.Context, id int) (*model.FieldDTO, error) {
	f, err := r.GetRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	return fieldToDTO(f), nil
}

func (r *FieldRepository) GetRaw(ctx context.Context, id int) (*model.Field, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT 
			id, 
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
		FROM fields WHERE id=$1
	`, id)
	var f model.Field
	if err := row.Scan(
		&f.ID,
		&f.CollectionID,
		&f.Name,
		&f.Label,
		&f.Type,
		&f.Required,
		&f.Unique,
		&f.Tag,
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

	return &f, nil
}

func (r *FieldRepository) Create(ctx context.Context, f *model.Field) (*model.FieldDTO, error) {
	row := r.DB.QueryRowContext(ctx, `
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
		RETURNING id
	`,
		f.CollectionID,
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
	if err := row.Scan(&f.ID); err != nil {
		return nil, err
	}
	return fieldToDTO(f), nil
}

func (r *FieldRepository) Update(ctx context.Context, f *model.Field) (*model.FieldDTO, error) {
	_, err := r.DB.ExecContext(ctx, `
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
		f.ID,
	)
	if err != nil {
		return nil, err
	}
	return fieldToDTO(f), nil
}

func (r *FieldRepository) Sort(ctx context.Context, ids []int) error {
	return dbutils.SortByIDs(ctx, r.DB, "fields", "order_index", ids)
}

func (r *FieldRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM fields WHERE id=$1`, id)
	return err
}
