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
  ('Sản phẩm - Xem', 'product.view'),
  ('Sản phẩm - Tạo', 'product.create'),
  ('Sản phẩm - Sửa', 'product.update'),
  ('Sản phẩm - Xoá', 'product.delete'),
  ('Sản phẩm - Tìm kiếm', 'product.search'),
	('Sản phẩm - Import', 'product.import'),
	('Sản phẩm - Export', 'product.export')
ON CONFLICT (permission_value)
DO UPDATE SET permission_name = EXCLUDED.permission_name;

-- ============================================
-- LINK ALL PERMISSIONS TO ADMIN ROLE
-- ============================================
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.permission_value IN (
  'product.view',
  'product.create',
  'product.update',
  'product.delete',
  'product.search',
	'product.import',
	'product.export'
)
WHERE r.role_name = 'admin'
ON CONFLICT DO NOTHING;
