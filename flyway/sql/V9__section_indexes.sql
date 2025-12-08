CREATE INDEX IF NOT EXISTS ix_section_id_not_deleted
  ON sections(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_section_with_department_id_not_deleted
  ON sections(department_id)
  WHERE deleted_at IS NULL;

-- custom fields
ALTER TABLE sections
  ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_sections_custom_fields_gin ON sections USING GIN (custom_fields);
