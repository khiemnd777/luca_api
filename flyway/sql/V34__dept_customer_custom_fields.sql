ALTER TABLE customers
  ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_customers_custom_fields_gin ON customers USING GIN (custom_fields);
