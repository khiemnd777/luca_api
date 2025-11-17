ALTER TABLE staffs
  ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_staffs_custom_fields_gin ON staffs USING GIN (custom_fields);
