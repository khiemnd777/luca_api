CREATE INDEX IF NOT EXISTS ix_material_id_not_deleted
  ON materials(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_material_code_not_deleted
  ON materials(code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_material_name_not_deleted
  ON materials(name)
  WHERE deleted_at IS NULL;
