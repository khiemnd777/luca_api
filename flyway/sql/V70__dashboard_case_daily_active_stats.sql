CREATE TABLE IF NOT EXISTS case_daily_active_stats (
  stat_date      DATE        NOT NULL,
  department_id  INT         NOT NULL,

  active_cases   INT         NOT NULL,

  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT pk_case_daily_active_stats
    PRIMARY KEY (stat_date, department_id)
);

CREATE INDEX IF NOT EXISTS idx_case_daily_active_stats_date
  ON case_daily_active_stats (stat_date);

CREATE INDEX IF NOT EXISTS idx_case_daily_active_stats_department
  ON case_daily_active_stats (department_id);
