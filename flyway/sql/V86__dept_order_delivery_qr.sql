CREATE TABLE IF NOT EXISTS order_delivery_qr_tokens (
  id SERIAL PRIMARY KEY,
  order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL,
  used BOOLEAN NOT NULL DEFAULT FALSE,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_order_delivery_qr_tokens_token_hash
  ON order_delivery_qr_tokens(token_hash);

CREATE INDEX IF NOT EXISTS ix_order_delivery_qr_tokens_order_used
  ON order_delivery_qr_tokens(order_id, used);

CREATE TABLE IF NOT EXISTS order_delivery_audit_logs (
  id SERIAL PRIMARY KEY,
  order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  qr_token_id INT REFERENCES order_delivery_qr_tokens(id) ON DELETE SET NULL,
  action TEXT NOT NULL,
  ip TEXT,
  user_agent TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_order_delivery_audit_logs_order_created
  ON order_delivery_audit_logs(order_id, created_at);

CREATE INDEX IF NOT EXISTS ix_order_delivery_audit_logs_qr_token_id
  ON order_delivery_audit_logs(qr_token_id);

CREATE INDEX IF NOT EXISTS ix_order_delivery_audit_logs_action
  ON order_delivery_audit_logs(action);

INSERT INTO permissions (permission_name, permission_value)
VALUES ('Đơn hàng - Xác nhận giao hàng', 'order.delivery')
ON CONFLICT (permission_value)
DO UPDATE SET permission_name = EXCLUDED.permission_name;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.permission_value IN ('order.delivery')
WHERE r.role_name = 'admin'
ON CONFLICT DO NOTHING;
