package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/khiemnd777/andy_api/modules/metadata/model"
)

type ImportFieldProfileRepository struct{ DB *sql.DB }

func NewImportFieldProfileRepository(db *sql.DB) *ImportFieldProfileRepository {
	return &ImportFieldProfileRepository{DB: db}
}

func (r *ImportFieldProfileRepository) List(ctx context.Context, scope string) ([]model.ImportFieldProfile, error) {
	scope = strings.TrimSpace(scope)

	query := `
		SELECT id, scope, code, name, pivot_field, permission, description, is_default
		FROM import_field_profiles
	`
	args := []any{}
	if scope != "" {
		query += ` WHERE scope = $1`
		args = append(args, scope)
	}
	query += ` ORDER BY scope ASC, code ASC`

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []model.ImportFieldProfile{}
	for rows.Next() {
		var (
			p    model.ImportFieldProfile
			desc sql.NullString
		)
		if err := rows.Scan(&p.ID, &p.Scope, &p.Code, &p.Name, &p.PivotField, &p.Permission, &desc, &p.IsDefault); err != nil {
			return nil, err
		}
		if desc.Valid {
			s := desc.String
			p.Description = &s
		}
		res = append(res, p)
	}
	return res, rows.Err()
}

func (r *ImportFieldProfileRepository) Get(ctx context.Context, id int) (*model.ImportFieldProfile, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, scope, code, name, pivot_field, permission, description, is_default
		FROM import_field_profiles
		WHERE id = $1
	`, id)

	var (
		p    model.ImportFieldProfile
		desc sql.NullString
	)
	if err := row.Scan(&p.ID, &p.Scope, &p.Code, &p.Name, &p.PivotField, &p.Permission, &desc, &p.IsDefault); err != nil {
		return nil, err
	}
	if desc.Valid {
		s := desc.String
		p.Description = &s
	}
	return &p, nil
}

func (r *ImportFieldProfileRepository) GetByScopeAndCode(ctx context.Context, scope, code string) (*model.ImportFieldProfile, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, scope, code, name, description, is_default, pivot_field, permission
		FROM import_field_profiles
		WHERE scope = $1 AND code = $2
	`, scope, code)

	var (
		p          model.ImportFieldProfile
		desc       sql.NullString
		pivot      sql.NullString
		permission sql.NullString
	)
	if err := row.Scan(&p.ID, &p.Scope, &p.Code, &p.Name, &desc, &p.IsDefault, &pivot, &permission); err != nil {
		return nil, err
	}
	if desc.Valid {
		s := desc.String
		p.Description = &s
	}
	if pivot.Valid {
		s := pivot.String
		p.PivotField = &s
	}
	if permission.Valid {
		s := permission.String
		p.Permission = &s
	}
	return &p, nil
}

func (r *ImportFieldProfileRepository) GetDefaultByScope(ctx context.Context, scope string) (*model.ImportFieldProfile, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, scope, code, name, description, is_default, pivot_field, permission
		FROM import_field_profiles
		WHERE scope = $1 AND is_default = TRUE
		LIMIT 1
	`, scope)

	var (
		p          model.ImportFieldProfile
		desc       sql.NullString
		pivot      sql.NullString
		permission sql.NullString
	)
	if err := row.Scan(&p.ID, &p.Scope, &p.Code, &p.Name, &desc, &p.IsDefault, &pivot, &permission); err != nil {
		return nil, err
	}
	if desc.Valid {
		s := desc.String
		p.Description = &s
	}
	if pivot.Valid {
		s := pivot.String
		p.PivotField = &s
	}
	if permission.Valid {
		s := permission.String
		p.Permission = &s
	}
	return &p, nil
}

func (r *ImportFieldProfileRepository) Create(ctx context.Context, p *model.ImportFieldProfile) (*model.ImportFieldProfile, error) {
	var desc sql.NullString
	if p.Description != nil && strings.TrimSpace(*p.Description) != "" {
		desc = sql.NullString{String: strings.TrimSpace(*p.Description), Valid: true}
	}

	row := r.DB.QueryRowContext(ctx, `
		INSERT INTO import_field_profiles (scope, code, name, pivot_field, permission, description, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, p.Scope, p.Code, p.Name, p.PivotField, p.Permission, desc, p.IsDefault)

	if err := row.Scan(&p.ID); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ImportFieldProfileRepository) Update(ctx context.Context, p *model.ImportFieldProfile) (*model.ImportFieldProfile, error) {
	var desc sql.NullString
	if p.Description != nil && strings.TrimSpace(*p.Description) != "" {
		desc = sql.NullString{String: strings.TrimSpace(*p.Description), Valid: true}
	}

	_, err := r.DB.ExecContext(ctx, `
		UPDATE import_field_profiles
		SET scope = $1, code = $2, name = $3, pivot_field = $4, permission = $5 , description = $6, is_default = $7
		WHERE id = $8
	`, p.Scope, p.Code, p.Name, p.PivotField, p.Permission, desc, p.IsDefault, p.ID)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ImportFieldProfileRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM import_field_profiles WHERE id = $1`, id)
	return err
}

func (r *ImportFieldProfileRepository) UnsetDefaultByScope(ctx context.Context, scope string) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE import_field_profiles
		SET is_default = FALSE
		WHERE scope = $1
	`, scope)
	return err
}
