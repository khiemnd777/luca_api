ALTER TABLE order_delivery_proofs
ADD COLUMN IF NOT EXISTS order_item_id BIGINT;

ALTER TABLE order_delivery_proofs
DROP CONSTRAINT IF EXISTS orderdeliveryproof_order_item_id;

DROP INDEX IF EXISTS orderdeliveryproof_order_item_id;
DROP INDEX IF EXISTS ux_order_delivery_proofs_order_item_id;

UPDATE order_delivery_proofs odp
SET order_item_id = latest_item.id
FROM (
  SELECT DISTINCT ON (oi.order_id)
    oi.order_id,
    oi.id
  FROM order_items oi
  WHERE oi.deleted_at IS NULL
  ORDER BY oi.order_id, oi.created_at DESC, oi.id DESC
) AS latest_item
WHERE odp.order_item_id IS NULL
  AND odp.order_id = latest_item.order_id;

ALTER TABLE order_delivery_proofs
ALTER COLUMN order_item_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS ix_order_delivery_proofs_order_item_id
  ON order_delivery_proofs(order_item_id);
