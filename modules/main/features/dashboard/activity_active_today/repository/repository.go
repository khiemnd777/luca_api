package repository

import (
	"context"
	"database/sql"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
)

type ActiveTodayRepository interface {
	ActiveToday(
		ctx context.Context,
		deptID int,
	) ([]*model.ActiveTodayItem, error)
}

type activeTodayRepository struct {
	db    *generated.Client
	sqlDB *sql.DB
	deps  *module.ModuleDeps[config.ModuleConfig]
}

func NewActiveTodayRepository(
	db *generated.Client,
	sqlDB *sql.DB,
	deps *module.ModuleDeps[config.ModuleConfig],
) ActiveTodayRepository {
	return &activeTodayRepository{
		db:    db,
		sqlDB: sqlDB,
		deps:  deps,
	}
}

func (r *activeTodayRepository) ActiveToday(
	ctx context.Context,
	deptID int,
) ([]*model.ActiveTodayItem, error) {

	const q = `
SELECT
	o.id,
	oi.code,
	o.dentist_name,
	o.patient_name,
	(oi.custom_fields->>'delivery_date')::timestamptz AS delivery_at,
	oi.created_at,
	(current_date - oi.created_at::date) AS age_days,
	oi.custom_fields->>'priority' AS priority
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE
	o.department_id = $1::INT
	AND o.deleted_at IS NULL
	AND oi.deleted_at IS NULL
	AND oi.custom_fields->>'status' IN (
		'received',
		'in_progress',
		'qc',
		'issue',
		'rework'
	)
ORDER BY
	oi.created_at ASC
LIMIT 5;
`

	rows, err := r.sqlDB.QueryContext(ctx, q, deptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.ActiveTodayItem

	for rows.Next() {
		var it model.ActiveTodayItem
		if err := rows.Scan(
			&it.ID,
			&it.Code,
			&it.Dentist,
			&it.Patient,
			&it.DeliveryAt,
			&it.CreatedAt,
			&it.AgeDays,
			&it.Priority,
		); err != nil {
			return nil, err
		}
		result = append(result, &it)
	}

	return result, rows.Err()
}
