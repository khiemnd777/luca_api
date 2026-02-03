CREATE TABLE IF NOT EXISTS sales_daily_stats (
  date               date          NOT NULL,
  department_id      int           NOT NULL,

  total_revenue      numeric(14,2) NOT NULL DEFAULT 0,
  order_items_count  int           NOT NULL DEFAULT 0,

  created_at         timestamptz   NOT NULL DEFAULT now(),
  updated_at         timestamptz   NOT NULL DEFAULT now(),

  PRIMARY KEY (date, department_id)
);

CREATE INDEX IF NOT EXISTS idx_sales_daily_stats_date
  ON sales_daily_stats(date);
