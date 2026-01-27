package repository

import (
	"context"
	"database/sql"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
)

type CaseStatusesRepository interface {
	CountByStatus(
		ctx context.Context,
		departmentID int,
	) ([]*model.CaseStatusCount, error)
}

type orderItemRepository struct {
	db    *generated.Client
	sqlDB *sql.DB
	deps  *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseStatusesRepository(
	db *generated.Client,
	sqlDB *sql.DB,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseStatusesRepository {
	return &orderItemRepository{
		db:    db,
		sqlDB: sqlDB,
		deps:  deps,
	}
}

func (r *orderItemRepository) CountByStatus(
	ctx context.Context,
	departmentID int,
) ([]*model.CaseStatusCount, error) {
	const q = `
SELECT
  oi.custom_fields->>'status' AS status,
  COUNT(*)                    AS count
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE
  o.department_id = $1
  AND oi.custom_fields->>'status' IN (
    'received',
    'in_progress',
    'qc',
    'issue',
    'rework'
  )
  AND (oi.custom_fields->>'delivery_date')::timestamptz >= date_trunc('day', now())
  AND (oi.custom_fields->>'delivery_date')::timestamptz <  date_trunc('day', now()) + interval '1 day'
GROUP BY
  oi.custom_fields->>'status';
`

	rows, err := r.sqlDB.QueryContext(ctx, q, departmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*model.CaseStatusCount, 0, 5)
	for rows.Next() {
		var rcd model.CaseStatusCount
		if err := rows.Scan(&rcd.Status, &rcd.Count); err != nil {
			return nil, err
		}
		res = append(res, &rcd)
	}

	return res, rows.Err()
}
