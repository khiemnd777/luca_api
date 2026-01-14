CREATE INDEX IF NOT EXISTS ix_order_items_id_not_deleted
  ON order_items(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_order_items_code_not_deleted
  ON order_items(code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_order_items_latest_by_code_original_not_deleted
  ON order_items(code_original, created_at)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_order_items_latest_by_order_id_not_deleted
  ON order_items(order_id, created_at)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_order_items_parent_order_id_not_deleted
  ON order_items(parent_item_id)
  WHERE deleted_at IS NULL;

-- order_item_products indexes
CREATE INDEX IF NOT EXISTS idx_order_item_products_order_id
ON order_item_products (order_id);

CREATE INDEX IF NOT EXISTS idx_order_item_products_order_id_order_item_id
ON order_item_products (order_id, order_item_id);

-- order_item_materials indexes
CREATE INDEX IF NOT EXISTS idx_order_item_materials_order_id
ON order_item_materials (order_id);

CREATE INDEX IF NOT EXISTS idx_order_item_materials_order_id_order_item_id
ON order_item_materials (order_id, order_item_id);

CREATE INDEX IF NOT EXISTS idx_oim_loaner_onloan_root_id
ON order_item_materials (id)
WHERE type = 'loaner'
  AND status IN ('on_loan', 'partial_returned')
  AND is_cloneable IS NULL;

CREATE INDEX IF NOT EXISTS idx_oim_order_item_id
  ON order_item_materials (order_item_id);

CREATE INDEX IF NOT EXISTS idx_oim_material_id
  ON order_item_materials (material_id);

