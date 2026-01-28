CREATE TABLE IF NOT EXISTS case_daily_remake_stats (
  stat_date        DATE        NOT NULL,
  department_id    INT         NOT NULL,

  completed_cases  INT         NOT NULL,
  remake_cases     INT         NOT NULL,

  remake_rate      NUMERIC GENERATED ALWAYS AS (
    CASE
      WHEN completed_cases > 0
      THEN remake_cases::numeric / completed_cases
      ELSE 0
    END
  ) STORED,

  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT pk_case_daily_remake_stats
    PRIMARY KEY (stat_date, department_id)
);

CREATE INDEX IF NOT EXISTS idx_case_daily_remake_stats_date
  ON case_daily_remake_stats (stat_date);

CREATE INDEX IF NOT EXISTS idx_case_daily_remake_stats_department
  ON case_daily_remake_stats (department_id);
