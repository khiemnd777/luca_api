ALTER TABLE processes
  ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_processes_custom_fields_gin ON processes USING GIN (custom_fields);
