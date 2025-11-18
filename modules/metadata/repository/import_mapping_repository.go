package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/khiemnd777/andy_api/modules/metadata/model"
)

type ImportFieldMappingRepository struct{ DB *sql.DB }

func NewImportFieldMappingRepository(db *sql.DB) *ImportFieldMappingRepository {
	return &ImportFieldMappingRepository{DB: db}
}

func (r *ImportFieldMappingRepository) ListByProfileID(ctx context.Context, profileID int) ([]model.ImportFieldMapping, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT
			id,
			profile_id,
			internal_kind,
			internal_path,
			internal_label,
			metadata_collection_slug,
			metadata_field_name,
			data_type,
			excel_header,
			excel_column,
			required,
			"unique",
			transform_hint
		FROM import_field_mappings
		WHERE profile_id = $1
		ORDER BY id ASC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []model.ImportFieldMapping{}
	for rows.Next() {
		var (
			m      model.ImportFieldMapping
			mcSlug sql.NullString
			mfName sql.NullString
			hNull  sql.NullString
			col    sql.NullInt32
			tHint  sql.NullString
		)
		if err := rows.Scan(
			&m.ID,
			&m.ProfileID,
			&m.InternalKind,
			&m.InternalPath,
			&m.InternalLabel,
			&mcSlug,
			&mfName,
			&m.DataType,
			&hNull,
			&col,
			&m.Required,
			&m.Unique,
			&tHint,
		); err != nil {
			return nil, err
		}
		if mcSlug.Valid {
			s := mcSlug.String
			m.MetadataCollectionSlug = &s
		}
		if mfName.Valid {
			s := mfName.String
			m.MetadataFieldName = &s
		}
		if hNull.Valid {
			s := hNull.String
			m.ExcelHeader = &s
		}
		if col.Valid {
			v := int(col.Int32)
			m.ExcelColumn = &v
		}
		if tHint.Valid {
			s := tHint.String
			m.TransformHint = &s
		}
		res = append(res, m)
	}
	return res, rows.Err()
}

func (r *ImportFieldMappingRepository) Get(ctx context.Context, id int) (*model.ImportFieldMapping, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT
			id,
			profile_id,
			internal_kind,
			internal_path,
			internal_label,
			metadata_collection_slug,
			metadata_field_name,
			data_type,
			excel_header,
			excel_column,
			required,
			"unique",
			transform_hint
		FROM import_field_mappings
		WHERE id = $1
	`, id)

	var (
		m      model.ImportFieldMapping
		mcSlug sql.NullString
		mfName sql.NullString
		hNull  sql.NullString
		col    sql.NullInt32
		tHint  sql.NullString
	)
	if err := row.Scan(
		&m.ID,
		&m.ProfileID,
		&m.InternalKind,
		&m.InternalPath,
		&m.InternalLabel,
		&mcSlug,
		&mfName,
		&m.DataType,
		&hNull,
		&col,
		&m.Required,
		&m.Unique,
		&tHint,
	); err != nil {
		return nil, err
	}
	if mcSlug.Valid {
		s := mcSlug.String
		m.MetadataCollectionSlug = &s
	}
	if mfName.Valid {
		s := mfName.String
		m.MetadataFieldName = &s
	}
	if hNull.Valid {
		s := hNull.String
		m.ExcelHeader = &s
	}
	if col.Valid {
		v := int(col.Int32)
		m.ExcelColumn = &v
	}
	if tHint.Valid {
		s := tHint.String
		m.TransformHint = &s
	}
	return &m, nil
}

func (r *ImportFieldMappingRepository) Create(ctx context.Context, m *model.ImportFieldMapping) (*model.ImportFieldMapping, error) {
	var (
		mcSlug sql.NullString
		mfName sql.NullString
		hNull  sql.NullString
		col    sql.NullInt32
		tHint  sql.NullString
	)

	if m.MetadataCollectionSlug != nil && strings.TrimSpace(*m.MetadataCollectionSlug) != "" {
		mcSlug = sql.NullString{String: strings.TrimSpace(*m.MetadataCollectionSlug), Valid: true}
	}
	if m.MetadataFieldName != nil && strings.TrimSpace(*m.MetadataFieldName) != "" {
		mfName = sql.NullString{String: strings.TrimSpace(*m.MetadataFieldName), Valid: true}
	}
	if m.ExcelHeader != nil && strings.TrimSpace(*m.ExcelHeader) != "" {
		hNull = sql.NullString{String: strings.TrimSpace(*m.ExcelHeader), Valid: true}
	}
	if m.ExcelColumn != nil && *m.ExcelColumn > 0 {
		col = sql.NullInt32{Int32: int32(*m.ExcelColumn), Valid: true}
	}
	if m.TransformHint != nil && strings.TrimSpace(*m.TransformHint) != "" {
		tHint = sql.NullString{String: strings.TrimSpace(*m.TransformHint), Valid: true}
	}

	row := r.DB.QueryRowContext(ctx, `
		INSERT INTO import_field_mappings (
			profile_id,
			internal_kind,
			internal_path,
			internal_label,
			metadata_collection_slug,
			metadata_field_name,
			data_type,
			excel_header,
			excel_column,
			required,
			"unique",
			transform_hint
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id
	`,
		m.ProfileID,
		m.InternalKind,
		m.InternalPath,
		m.InternalLabel,
		mcSlug,
		mfName,
		m.DataType,
		hNull,
		col,
		m.Required,
		m.Unique,
		tHint,
	)

	if err := row.Scan(&m.ID); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *ImportFieldMappingRepository) Update(ctx context.Context, m *model.ImportFieldMapping) (*model.ImportFieldMapping, error) {
	var (
		mcSlug sql.NullString
		mfName sql.NullString
		hNull  sql.NullString
		col    sql.NullInt32
		tHint  sql.NullString
	)

	if m.MetadataCollectionSlug != nil && strings.TrimSpace(*m.MetadataCollectionSlug) != "" {
		mcSlug = sql.NullString{String: strings.TrimSpace(*m.MetadataCollectionSlug), Valid: true}
	}
	if m.MetadataFieldName != nil && strings.TrimSpace(*m.MetadataFieldName) != "" {
		mfName = sql.NullString{String: strings.TrimSpace(*m.MetadataFieldName), Valid: true}
	}
	if m.ExcelHeader != nil && strings.TrimSpace(*m.ExcelHeader) != "" {
		hNull = sql.NullString{String: strings.TrimSpace(*m.ExcelHeader), Valid: true}
	}
	if m.ExcelColumn != nil && *m.ExcelColumn > 0 {
		col = sql.NullInt32{Int32: int32(*m.ExcelColumn), Valid: true}
	}
	if m.TransformHint != nil && strings.TrimSpace(*m.TransformHint) != "" {
		tHint = sql.NullString{String: strings.TrimSpace(*m.TransformHint), Valid: true}
	}

	_, err := r.DB.ExecContext(ctx, `
		UPDATE import_field_mappings
		SET
			profile_id = $1,
			internal_kind = $2,
			internal_path = $3,
			internal_label = $4,
			metadata_collection_slug = $5,
			metadata_field_name = $6,
			data_type = $7,
			excel_header = $8,
			excel_column = $9,
			required = $10,
			"unique" = $11,
			transform_hint = $12
		WHERE id = $13
	`,
		m.ProfileID,
		m.InternalKind,
		m.InternalPath,
		m.InternalLabel,
		mcSlug,
		mfName,
		m.DataType,
		hNull,
		col,
		m.Required,
		m.Unique,
		tHint,
		m.ID,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *ImportFieldMappingRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM import_field_mappings WHERE id = $1`, id)
	return err
}
