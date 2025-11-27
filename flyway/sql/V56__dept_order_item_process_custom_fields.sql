ALTER TABLE order_item_processes
  ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_order_item_process_custom_fields_gin ON order_item_processes USING GIN (custom_fields);
