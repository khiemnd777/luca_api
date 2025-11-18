CREATE INDEX IF NOT EXISTS ix_supplier_id_not_deleted
  ON suppliers(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_supplier_code_not_deleted
  ON suppliers(code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_supplier_name_not_deleted
  ON suppliers(name)
  WHERE deleted_at IS NULL;
