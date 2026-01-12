CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id);
CREATE INDEX IF NOT EXISTS ix_orders_id_not_deleted
  ON orders(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_orders_code_not_deleted
  ON orders(code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_orders_customer_not_deleted
  ON orders(customer_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_orders_in_progress_delivery_date
ON orders (delivery_date DESC)
WHERE
  deleted_at IS NULL
  AND status_latest = 'in_progress';

CREATE INDEX IF NOT EXISTS idx_orders_received_created_at
ON orders (created_at DESC)
WHERE
  deleted_at IS NULL
  AND status_latest = 'received';

CREATE INDEX IF NOT EXISTS idx_orders_completed_updated_at
ON orders (updated_at DESC)
WHERE
  deleted_at IS NULL
  AND status_latest = 'completed';
