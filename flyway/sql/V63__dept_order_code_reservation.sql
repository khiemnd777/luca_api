CREATE TABLE IF NOT EXISTS order_code_counters (
  period CHAR(4) PRIMARY KEY,  -- 'MMYY' ví dụ '1225'
  last_seq INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS order_code_reservations (
  order_code VARCHAR(32) PRIMARY KEY,
  period CHAR(4) NOT NULL,
  seq INTEGER NOT NULL,
  status VARCHAR(16) NOT NULL,      -- reserved | used | expired | cancelled
  reserved_at TIMESTAMP NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  used_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_order_code_resv_status_expires
  ON order_code_reservations(status, expires_at);

CREATE INDEX IF NOT EXISTS idx_order_code_resv_period_seq
  ON order_code_reservations(period, seq);
