CREATE INDEX IF NOT EXISTS ix_category_id_not_deleted
  ON categories(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_category_name_not_deleted
  ON categories(name)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_category_processes_category_order
ON category_processes(category_id, display_order);
