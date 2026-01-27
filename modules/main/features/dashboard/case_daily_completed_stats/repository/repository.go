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

type CaseDailyCompletedStatsRepository interface {
	UpsertOne(
		ctx context.Context,
		completedAt time.Time,
		departmentID int,
	) error

	RebuildRange(
		ctx context.Context,
		fromDate time.Time,
		toDate time.Time,
	) error

	CompletedCases(
		ctx context.Context,
		departmentID *int, // nil = all departments
		fromDate time.Time,
		toDate time.Time,
		previousFrom time.Time,
		previousTo time.Time,
	) (*model.CompletedCasesResult, error)
}

type caseDailyCompletedStatsRepository struct {
	db    *generated.Client
	sqlDB *sql.DB
	deps  *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseDailyCompletedStatsRepository(
	db *generated.Client,
	sqlDB *sql.DB,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseDailyCompletedStatsRepository {
	return &caseDailyCompletedStatsRepository{
		db:    db,
		sqlDB: sqlDB,
		deps:  deps,
	}
}

func (r *caseDailyCompletedStatsRepository) UpsertOne(
	ctx context.Context,
	statDate time.Time,
	departmentID int,
) error {

	const q = `
INSERT INTO case_daily_active_stats (
  stat_date,
  active_cases
)
SELECT
  $1 AS stat_date,
  COUNT(*) AS active_cases
FROM order_items oi
WHERE
  oi.custom_fields->>'status' IN (
    'received',
    'in_progress',
    'qc',
    'issue',
    'rework'
  )
ON CONFLICT (stat_date) DO UPDATE
SET
  active_cases = EXCLUDED.active_cases,
  updated_at = now();
`

	_, err := r.sqlDB.ExecContext(
		ctx,
		q,
		statDate.UTC(),
	)

	return err
}

func (r *caseDailyCompletedStatsRepository) RebuildRange(
	ctx context.Context,
	fromDate time.Time,
	toDate time.Time,
) error {
	const q = `
INSERT INTO case_daily_completed_stats (
  stat_date,
  department_id,
  completed_cases
)
SELECT
  DATE(completed_at),
  department_id,
  COUNT(*)
FROM cases
WHERE
  completed_at >= $1
  AND completed_at <  $2
GROUP BY
  DATE(completed_at),
  department_id
ON CONFLICT (stat_date, department_id) DO UPDATE
SET
  completed_cases = EXCLUDED.completed_cases,
  updated_at = now();
`
	_, err := r.sqlDB.ExecContext(
		ctx,
		q,
		fromDate.UTC(),
		toDate.UTC(),
	)
	return err
}

func (r *caseDailyCompletedStatsRepository) CompletedCases(
	ctx context.Context,
	departmentID *int, // nil = all departments
	fromDate time.Time,
	toDate time.Time,
	previousFrom time.Time,
	previousTo time.Time,
) (*model.CompletedCasesResult, error) {

	const q = `
WITH current_period AS (
  SELECT COALESCE(SUM(completed_cases), 0) AS value
  FROM case_daily_completed_stats
  WHERE
    stat_date >= $1
    AND stat_date <  $2
    AND ($3::INT IS NULL OR department_id = $3::INT)
),
previous_period AS (
  SELECT COALESCE(SUM(completed_cases), 0) AS value
  FROM case_daily_completed_stats
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

	var res model.CompletedCasesResult

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
