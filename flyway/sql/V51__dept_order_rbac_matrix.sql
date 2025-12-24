-- ============================================
-- RBAC PERMISSIONS + ADMIN ROLE UPSERT SCRIPT
-- ============================================

-- 1. Ensure role "admin" exists
INSERT INTO roles (role_name)
VALUES ('admin')
ON CONFLICT (role_name)
DO UPDATE SET role_name = EXCLUDED.role_name;

-- ============================================
-- PERMISSIONS UPSERT
-- ============================================
INSERT INTO permissions (permission_name, permission_value)
VALUES
--  Order
  ('Đơn hàng - Xem', 'order.view'),
  ('Đơn hàng - Tạo', 'order.create'),
  ('Đơn hàng - Sửa', 'order.update'),
  ('Đơn hàng - Xoá', 'order.delete'),
  ('Đơn hàng - Tìm kiếm', 'order.search'),
	('Đơn hàng - Import', 'order.import'),
	('Đơn hàng - Export', 'order.export'),
--  Order development
  ('Đơn hàng - Gia công', 'order.development')

ON CONFLICT (permission_value)
DO UPDATE SET permission_name = EXCLUDED.permission_name;

-- ============================================
-- LINK ALL PERMISSIONS TO ADMIN ROLE
-- ============================================
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.permission_value IN (
  'order.view',
  'order.create',
  'order.update',
  'order.delete',
  'order.search',
	'order.import',
	'order.export',
  'order.development'
)
WHERE r.role_name = 'admin'
ON CONFLICT DO NOTHING;
