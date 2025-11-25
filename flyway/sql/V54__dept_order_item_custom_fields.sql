ALTER TABLE order_items
  ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_order_item_custom_fields_gin ON order_items USING GIN (custom_fields);
