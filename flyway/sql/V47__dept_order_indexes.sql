CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
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