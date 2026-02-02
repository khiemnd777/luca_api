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

type CaseDailySalesStatsRepository interface {
	UpsertOne(
		ctx context.Context,
		from time.Time,
		to time.Time,
	) error

	Summary(
		ctx context.Context,
		deptID int,
		from time.Time,
		to time.Time,
		prevFrom time.Time,
		prevTo time.Time,
	) (*model.SalesSummary, error)

	Daily(
		ctx context.Context,
		deptID int,
		from time.Time,
		to time.Time,
	) ([]*model.SalesDailyItem, error)
}

type caseDailySalesStatsRepository struct {
	db    *generated.Client
	sqlDB *sql.DB
	deps  *module.ModuleDeps[config.ModuleConfig]
}

func NewCaseDailySalesStatsRepository(
	db *generated.Client,
	sqlDB *sql.DB,
	deps *module.ModuleDeps[config.ModuleConfig],
) CaseDailySalesStatsRepository {
	return &caseDailySalesStatsRepository{
		db:    db,
		sqlDB: sqlDB,
		deps:  deps,
	}
}

func (r *caseDailySalesStatsRepository) UpsertOne(
	ctx context.Context,
	from time.Time,
	to time.Time,
) error {
	const q = `
INSERT INTO sales_daily_stats (
  date,
  department_id,
  total_revenue,
  order_items_count
)
SELECT
  date_trunc('day', o.created_at)::date AS date,
  o.department_id,
  SUM(o.total_price)                    AS total_revenue,
  COUNT(oi.id)                          AS order_items_count
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE
  oi.custom_fields->>'status' = 'completed'
  AND o.created_at >= $1
  AND o.created_at <  $2
GROUP BY date, o.department_id
ON CONFLICT (date, department_id)
DO UPDATE SET
  total_revenue     = EXCLUDED.total_revenue,
  order_items_count = EXCLUDED.order_items_count,
  updated_at        = now();
`

	_, err := r.sqlDB.ExecContext(
		ctx,
		q,
		from.UTC(),
		to.UTC(),
	)

	return err
}

func (r *caseDailySalesStatsRepository) Summary(
	ctx context.Context,
	deptID int,
	from time.Time,
	to time.Time,
	prevFrom time.Time,
	prevTo time.Time,
) (*model.SalesSummary, error) {
	const q = `
WITH current_period AS (
  SELECT
    COALESCE(SUM(total_revenue), 0)::float8     AS total_revenue,
    COALESCE(SUM(order_items_count), 0)::int   AS order_items_count
  FROM sales_daily_stats
  WHERE
    date >= $1
    AND date <  $2
    AND department_id = $3
),
previous_period AS (
  SELECT
    COALESCE(SUM(total_revenue), 0)::float8 AS prev_revenue
  FROM sales_daily_stats
  WHERE
    date >= $4
    AND date <  $5
    AND department_id = $3
)
SELECT
  c.total_revenue,
  c.order_items_count,
  p.prev_revenue
FROM current_period c
CROSS JOIN previous_period p;
`

	var res model.SalesSummary

	err := r.sqlDB.QueryRowContext(
		ctx,
		q,
		from,
		to,
		deptID,
		prevFrom,
		prevTo,
	).Scan(&res.TotalRevenue, &res.OrderItemsCount, &res.PrevRevenue)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *caseDailySalesStatsRepository) Daily(
	ctx context.Context,
	deptID int,
	from time.Time,
	to time.Time,
) ([]*model.SalesDailyItem, error) {
	const q = `
SELECT
  date,
  COALESCE(SUM(total_revenue), 0)::float8 AS revenue
FROM sales_daily_stats
WHERE
  date >= $1
  AND date <  $2
  AND department_id = $3
GROUP BY date
ORDER BY date;
`

	rows, err := r.sqlDB.QueryContext(ctx, q, from, to, deptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*model.SalesDailyItem{}
	for rows.Next() {
		var item model.SalesDailyItem
		if err := rows.Scan(&item.Date, &item.Revenue); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
