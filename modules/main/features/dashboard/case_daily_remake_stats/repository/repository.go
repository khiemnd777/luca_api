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

type CaseDailyRemakeStatsRepository interface {
	UpsertOne(
		ctx context.Context,
		completedAt time.Time,
		departmentID int,
		isRemake bool,
	) error

	RebuildRange(
		ctx context.Context,
		fromDate time.Time,
		toDate time.Time,
	) error

	AvgRemakeRate(
		ctx context.Context,
		departmentID *int, // nil = all departments
		fromDate time.Time,
		toDate time.Time,
		previousFrom time.Time,
		previousTo time.Time,
	) (*model.AvgRemakeResult, error)
}

type caseDailyRemakeStatsRepository struct {
	db    *generated.Client
	sqlDB *sql.DB
	deps  *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseDailyRemakeStatsRepository(
	db *generated.Client,
	sqlDB *sql.DB,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseDailyRemakeStatsRepository {
	return &caseDailyRemakeStatsRepository{
		db:    db,
		sqlDB: sqlDB,
		deps:  deps,
	}
}

func (r *caseDailyRemakeStatsRepository) UpsertOne(
	ctx context.Context,
	completedAt time.Time,
	departmentID int,
	isRemake bool,
) error {
	const q = `
INSERT INTO case_daily_remake_stats (
  stat_date,
  department_id,
  completed_cases,
  remake_cases
)
VALUES (
  $1,
  $2,
  1,
  CASE WHEN $3 THEN 1 ELSE 0 END
)
ON CONFLICT (stat_date, department_id) DO UPDATE
SET
  completed_cases = case_daily_remake_stats.completed_cases + 1,
  remake_cases    = case_daily_remake_stats.remake_cases + CASE WHEN $3 THEN 1 ELSE 0 END,
  updated_at      = now();
`

	_, err := r.sqlDB.ExecContext(
		ctx,
		q,
		completedAt.UTC().Truncate(24*time.Hour),
		departmentID,
		isRemake,
	)

	return err
}

func (r *caseDailyRemakeStatsRepository) RebuildRange(
	ctx context.Context,
	fromDate time.Time,
	toDate time.Time,
) error {
	const q = `
INSERT INTO case_daily_remake_stats (
  stat_date,
  department_id,
  completed_cases,
  remake_cases
)
SELECT
  DATE(completed_at)                                  AS stat_date,
  department_id,
  COUNT(*)                                           AS completed_cases,
  COUNT(*) FILTER (WHERE is_remake = true)           AS remake_cases
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
  remake_cases    = EXCLUDED.remake_cases,
  updated_at      = now();
`

	_, err := r.sqlDB.ExecContext(
		ctx,
		q,
		fromDate.UTC(),
		toDate.UTC(),
	)

	return err
}

func (r *caseDailyRemakeStatsRepository) AvgRemakeRate(
	ctx context.Context,
	departmentID *int,
	fromDate time.Time,
	toDate time.Time,
	previousFrom time.Time,
	previousTo time.Time,
) (*model.AvgRemakeResult, error) {

	const q = `
WITH current_period AS (
  SELECT
    SUM(remake_cases)::numeric / NULLIF(SUM(completed_cases), 0) AS rate
  FROM case_daily_remake_stats
  WHERE
    stat_date >= $1
    AND stat_date <  $2
    AND ($3 IS NULL OR department_id = $3)
),
previous_period AS (
  SELECT
    SUM(remake_cases)::numeric / NULLIF(SUM(completed_cases), 0) AS rate
  FROM case_daily_remake_stats
  WHERE
    stat_date >= $4
    AND stat_date <  $5
    AND ($3 IS NULL OR department_id = $3)
)
SELECT
  COALESCE(c.rate, 0)             AS rate,
  COALESCE(c.rate - p.rate, 0)    AS delta_rate
FROM current_period c
CROSS JOIN previous_period p;
`

	var res model.AvgRemakeResult

	err := r.sqlDB.QueryRowContext(
		ctx,
		q,
		fromDate,
		toDate,
		departmentID,
		previousFrom,
		previousTo,
	).Scan(&res.Rate, &res.DeltaRate)

	if err != nil {
		return nil, err
	}

	return &res, nil
}
