package repository

import (
	"context"
	"database/sql"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
)

type DueTodayRepository interface {
	DueToday(
		ctx context.Context,
		deptID int,
	) ([]*model.DueTodayItem, error)
}

type dueTodayRepository struct {
	db    *generated.Client
	sqlDB *sql.DB
	deps  *module.ModuleDeps[config.ModuleConfig]
}

func NewDueTodayRepository(
	db *generated.Client,
	sqlDB *sql.DB,
	deps *module.ModuleDeps[config.ModuleConfig],
) DueTodayRepository {
	return &dueTodayRepository{
		db:    db,
		sqlDB: sqlDB,
		deps:  deps,
	}
}

func (r *dueTodayRepository) DueToday(
	ctx context.Context,
	deptID int,
) ([]*model.DueTodayItem, error) {

	const q = `
SELECT
	o.id,
  oi.code,
  o.dentist_name,
  o.patient_name,
  (oi.custom_fields->>'delivery_date')::timestamptz AS delivery_at,
  oi.custom_fields->>'priority'       AS priority
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE
	o.department_id=$1::INT
  AND (oi.custom_fields->>'delivery_date')::timestamptz >= date_trunc('day', now())
  AND (oi.custom_fields->>'delivery_date')::timestamptz <  date_trunc('day', now()) + interval '1 day'
  AND oi.custom_fields->>'status' IN (
    'received',
    'in_progress',
    'qc',
    'issue',
    'rework'
  )
ORDER BY
  delivery_at ASC;
`

	rows, err := r.sqlDB.QueryContext(ctx, q, deptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.DueTodayItem

	for rows.Next() {
		var it model.DueTodayItem
		if err := rows.Scan(
			&it.ID,
			&it.Code,
			&it.Dentist,
			&it.Patient,
			&it.DeliveryAt,
			&it.Priority,
		); err != nil {
			return nil, err
		}
		result = append(result, &it)
	}

	return result, rows.Err()
}
