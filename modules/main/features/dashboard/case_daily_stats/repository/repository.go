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

type CaseDailyStatsRepository interface {
	UpsertOne(
		ctx context.Context,
		completedAt time.Time,
		departmentID int,
		turnaroundSec int64,
	) error

	RebuildRange(
		ctx context.Context,
		fromDate time.Time,
		toDate time.Time,
	) error

	AvgTurnaround(
		ctx context.Context,
		departmentID *int, // nil = all departments
		fromDate time.Time,
		toDate time.Time,
		previousFrom time.Time,
		previousTo time.Time,
	) (*model.AvgTurnaroundResult, error)
}

type caseDailyStatsRepository struct {
	db    *generated.Client
	sqlDB *sql.DB
	deps  *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseDailyStatsRepository(
	db *generated.Client,
	sqlDB *sql.DB,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseDailyStatsRepository {
	return &caseDailyStatsRepository{
		db:    db,
		sqlDB: sqlDB,
		deps:  deps,
	}
}

func (r *caseDailyStatsRepository) UpsertOne(
	ctx context.Context,
	completedAt time.Time,
	departmentID int,
	turnaroundSec int64,
) error {
	const q = `
INSERT INTO case_daily_stats (
  stat_date,
  department_id,
  completed_cases,
  total_turnaround_sec
)
VALUES (
  $1,
  $2,
  1,
  $3
)
ON CONFLICT (stat_date, department_id) DO UPDATE
SET
  completed_cases      = case_daily_stats.completed_cases + 1,
  total_turnaround_sec = case_daily_stats.total_turnaround_sec + EXCLUDED.total_turnaround_sec,
  updated_at           = now();
`

	_, err := r.sqlDB.ExecContext(
		ctx,
		q,
		completedAt.UTC().Truncate(24*time.Hour),
		departmentID,
		turnaroundSec,
	)

	return err
}

func (r *caseDailyStatsRepository) RebuildRange(
	ctx context.Context,
	fromDate time.Time,
	toDate time.Time,
) error {
	const q = `
INSERT INTO case_daily_stats (
  stat_date,
  department_id,
  completed_cases,
  total_turnaround_sec
)
SELECT
  DATE(completed_at)                     AS stat_date,
  department_id,
  COUNT(*)                               AS completed_cases,
  SUM(EXTRACT(EPOCH FROM (completed_at - received_at)))::bigint
FROM cases
WHERE
  completed_at >= $1
  AND completed_at <  $2
  AND completed_at IS NOT NULL
  AND received_at IS NOT NULL
GROUP BY
  DATE(completed_at),
  department_id
ON CONFLICT (stat_date, department_id) DO UPDATE
SET
  completed_cases      = EXCLUDED.completed_cases,
  total_turnaround_sec = EXCLUDED.total_turnaround_sec,
  updated_at           = now();
`

	_, err := r.sqlDB.ExecContext(
		ctx,
		q,
		fromDate.UTC(),
		toDate.UTC(),
	)

	return err
}

func (r *caseDailyStatsRepository) AvgTurnaround(
	ctx context.Context,
	departmentID *int,
	fromDate time.Time,
	toDate time.Time,
	previousFrom time.Time,
	previousTo time.Time,
) (*model.AvgTurnaroundResult, error) {

	const q = `
WITH current_period AS (
  SELECT
    SUM(total_turnaround_sec)::numeric
      / NULLIF(SUM(completed_cases), 0) AS avg_sec
  FROM case_daily_stats
  WHERE
    stat_date >= $1
    AND stat_date <  $2
    AND ($3 IS NULL OR department_id = $3)
),
previous_period AS (
  SELECT
    SUM(total_turnaround_sec)::numeric
      / NULLIF(SUM(completed_cases), 0) AS avg_sec
  FROM case_daily_stats
  WHERE
    stat_date >= $4
    AND stat_date <  $5
    AND ($3 IS NULL OR department_id = $3)
)
SELECT
  COALESCE(c.avg_sec / 86400, 0)                   AS avg_days,
  COALESCE((c.avg_sec - p.avg_sec) / 86400, 0)     AS delta_days
FROM current_period c
CROSS JOIN previous_period p;
`

	var res model.AvgTurnaroundResult

	err := r.sqlDB.QueryRowContext(
		ctx,
		q,
		fromDate,
		toDate,
		departmentID, // $3
		previousFrom,
		previousTo,
	).Scan(&res.AvgDays, &res.DeltaDays)

	if err != nil {
		return nil, err
	}

	return &res, nil
}
