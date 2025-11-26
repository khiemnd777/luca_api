CREATE INDEX IF NOT EXISTS ix_order_items_id_not_deleted
  ON order_items(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_order_items_code_not_deleted
  ON order_items(code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_order_items_latest_by_code_original_not_deleted
  ON order_items(code_original, created_at)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_order_items_latest_by_order_id_not_deleted
  ON order_items(order_id, created_at)
  WHERE deleted_at IS NULL;
