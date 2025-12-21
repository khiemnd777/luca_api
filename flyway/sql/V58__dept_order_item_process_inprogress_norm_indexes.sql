CREATE INDEX IF NOT EXISTS ix_order_item_process_in_progresses_order_item_id_must_completed_at
  ON order_item_process_in_progresses(order_item_id, created_at)
  WHERE completed_at IS NOT NULL;
