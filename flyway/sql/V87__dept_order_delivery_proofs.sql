CREATE TABLE IF NOT EXISTS order_delivery_proofs (
  id SERIAL PRIMARY KEY,
  order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  qr_token_id INT NOT NULL REFERENCES order_delivery_qr_tokens(id) ON DELETE CASCADE,
  image_url TEXT NOT NULL,
  image_size BIGINT NOT NULL,
  image_mime_type TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_order_delivery_proofs_order_id
  ON order_delivery_proofs(order_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_order_delivery_proofs_qr_token_id
  ON order_delivery_proofs(qr_token_id);
