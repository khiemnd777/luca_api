package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
)

type CaseDailyActiveStatsRepository interface {
	UpsertOne(
		ctx context.Context,
		statDate time.Time,
		departmentID int,
	) error

	ActiveCases(
		ctx context.Context,
		departmentID *int,
		fromDate time.Time,
		toDate time.Time,
		previousFrom time.Time,
		previousTo time.Time,
	) (*model.ActiveCasesResult, error)
}

type caseDailyActiveStatsRepository struct {
	db    *generated.Client
	sqlDB *sql.DB
	deps  *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseDailyActiveStatsRepository(
	db *generated.Client,
	sqlDB *sql.DB,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseDailyActiveStatsRepository {
	return &caseDailyActiveStatsRepository{
		db:    db,
		sqlDB: sqlDB,
		deps:  deps,
	}
}

func (r *caseDailyActiveStatsRepository) UpsertOne(
	ctx context.Context,
	statDate time.Time,
	departmentID int,
) error {

	const q = `
INSERT INTO case_daily_active_stats (
  stat_date,
  department_id,
  active_cases
)
SELECT
  $1 AS stat_date,
  $2 AS department_id,
  COUNT(*) AS active_cases
FROM (
  SELECT 1
  FROM order_items oi
  WHERE
    oi.custom_fields->>'status' IN (
      'received',
      'in_progress',
      'qc',
      'issue',
      'rework'
    )
) t
ON CONFLICT (stat_date, department_id) DO UPDATE
SET
  active_cases = EXCLUDED.active_cases,
  updated_at = now();
`

	_, err := r.sqlDB.ExecContext(
		ctx,
		q,
		statDate.UTC(),
		departmentID,
	)

	return err
}

func (r *caseDailyActiveStatsRepository) ActiveCases(
	ctx context.Context,
	departmentID *int,
	fromDate time.Time,
	toDate time.Time,
	previousFrom time.Time,
	previousTo time.Time,
) (*model.ActiveCasesResult, error) {

	const q = `
WITH current_period AS (
  SELECT
    COALESCE(SUM(active_cases), 0) AS value
  FROM case_daily_active_stats
  WHERE
    stat_date >= $1
    AND stat_date <  $2
    AND ($3::INT IS NULL OR department_id = $3::INT)
),
previous_period AS (
  SELECT
    COALESCE(SUM(active_cases), 0) AS value
  FROM case_daily_active_stats
  WHERE
    stat_date >= $4
    AND stat_date <  $5
    AND ($3::INT IS NULL OR department_id = $3::INT)
)
SELECT
  c.value AS value,
  (c.value - p.value) AS delta
FROM current_period c
CROSS JOIN previous_period p;
`

	var res model.ActiveCasesResult

	err := r.sqlDB.QueryRowContext(
		ctx,
		q,
		fromDate,
		toDate,
		departmentID,
		previousFrom,
		previousTo,
	).Scan(&res.Value, &res.Delta)

	if err != nil {
		return nil, err
	}

	return &res, nil
}
