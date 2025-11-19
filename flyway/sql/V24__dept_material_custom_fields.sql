ALTER TABLE materials
  ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_materials_custom_fields_gin ON materials USING GIN (custom_fields);
