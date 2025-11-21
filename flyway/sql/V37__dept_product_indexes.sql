CREATE INDEX IF NOT EXISTS ix_product_id_not_deleted
  ON products(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_product_code_not_deleted
  ON products(code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_product_name_not_deleted
  ON products(name)
  WHERE deleted_at IS NULL;
