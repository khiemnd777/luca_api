CREATE INDEX IF NOT EXISTS ix_department_id_not_deleted
  ON departments(slug)
  WHERE deleted = FALSE;
