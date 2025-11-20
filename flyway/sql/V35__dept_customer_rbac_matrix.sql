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
  ('Khách hàng - Xem', 'customer.view'),
  ('Khách hàng - Tạo', 'customer.create'),
  ('Khách hàng - Sửa', 'customer.update'),
  ('Khách hàng - Xoá', 'customer.delete'),
  ('Khách hàng - Tìm kiếm', 'customer.search'),
	('Khách hàng - Import', 'customer.import'),
	('Khách hàng - Export', 'customer.export')
ON CONFLICT (permission_value)
DO UPDATE SET permission_name = EXCLUDED.permission_name;

-- ============================================
-- LINK ALL PERMISSIONS TO ADMIN ROLE
-- ============================================
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.permission_value IN (
  'customer.view',
  'customer.create',
  'customer.update',
  'customer.delete',
  'customer.search',
	'customer.import',
	'customer.export'
)
WHERE r.role_name = 'admin'
ON CONFLICT DO NOTHING;
