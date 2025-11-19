ALTER TABLE suppliers
  ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_suppliers_custom_fields_gin ON suppliers USING GIN (custom_fields);
