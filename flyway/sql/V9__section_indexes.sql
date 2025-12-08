CREATE INDEX IF NOT EXISTS ix_section_id_not_deleted
  ON sections(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_section_with_department_id_not_deleted
  ON sections(department_id)
  WHERE deleted_at IS NULL;
