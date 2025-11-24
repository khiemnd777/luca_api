ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_order_custom_fields_gin ON orders USING GIN (custom_fields);
