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
		departmentID *int,
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
  $1::date,
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
		completedAt,
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

	tx, err := r.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const deleteQ = `
DELETE FROM case_daily_stats
WHERE
  stat_date >= $1::date
  AND stat_date <  $2::date;
`
	if _, err := tx.ExecContext(ctx, deleteQ, fromDate, toDate); err != nil {
		return err
	}

	const insertQ = `
INSERT INTO case_daily_stats (
  stat_date,
  department_id,
  completed_cases,
  total_turnaround_sec
)
SELECT
  oi.completed_at::date AS stat_date,
  o.department_id,
  COUNT(*) AS completed_cases,
  COALESCE(
    SUM(EXTRACT(EPOCH FROM (oi.completed_at - oi.created_at))),
    0
  )::bigint AS total_turnaround_sec
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE
  oi.completed_at >= $1
  AND oi.completed_at <  $2
  AND oi.completed_at IS NOT NULL
  AND oi.created_at IS NOT NULL
  AND oi.custom_fields->>'status' = 'completed'
  AND oi.deleted_at IS NULL
  AND o.deleted_at IS NULL
GROUP BY
  stat_date,
  o.department_id;
`
	if _, err := tx.ExecContext(ctx, insertQ, fromDate, toDate); err != nil {
		return err
	}

	return tx.Commit()
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
    stat_date >= $1::date
    AND stat_date <= $2::date
    AND ($3::INT IS NULL OR department_id = $3::INT)
),
previous_period AS (
  SELECT
    SUM(total_turnaround_sec)::numeric
      / NULLIF(SUM(completed_cases), 0) AS avg_sec
  FROM case_daily_stats
  WHERE
    stat_date >= $4::date
    AND stat_date <= $5::date
    AND ($3::INT IS NULL OR department_id = $3::INT)
)
SELECT
  COALESCE(c.avg_sec / 86400, 0)               AS avg_days,
  COALESCE((c.avg_sec - p.avg_sec) / 86400, 0) AS delta_days
FROM current_period c
CROSS JOIN previous_period p;
`

	var res model.AvgTurnaroundResult

	err := r.sqlDB.QueryRowContext(
		ctx,
		q,
		fromDate,
		toDate,
		departmentID,
		previousFrom,
		previousTo,
	).Scan(&res.AvgDays, &res.DeltaDays)

	if err != nil {
		return nil, err
	}

	return &res, nil
}
