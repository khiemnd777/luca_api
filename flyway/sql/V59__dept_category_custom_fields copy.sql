ALTER TABLE categories
  ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_categories_custom_fields_gin ON categories USING GIN (custom_fields);
