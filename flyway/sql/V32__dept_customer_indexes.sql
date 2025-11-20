CREATE INDEX IF NOT EXISTS ix_customer_id_not_deleted
  ON customers(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_customer_code_not_deleted
  ON customers(code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_customer_name_not_deleted
  ON customers(name)
  WHERE deleted_at IS NULL;
