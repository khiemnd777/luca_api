CREATE INDEX IF NOT EXISTS ix_departments_parent_id_not_deleted
  ON departments(parent_id)
  WHERE deleted = FALSE;

CREATE INDEX IF NOT EXISTS ix_department_id_not_deleted
  ON departments(id)
  WHERE deleted = FALSE;
