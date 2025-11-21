ALTER TABLE products
  ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_products_custom_fields_gin ON products USING GIN (custom_fields);
