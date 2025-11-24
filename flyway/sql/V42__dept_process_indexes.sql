CREATE INDEX IF NOT EXISTS ix_process_id_not_deleted
  ON processes(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_process_code_not_deleted
  ON processes(code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_process_name_not_deleted
  ON processes(name)
  WHERE deleted_at IS NULL;
