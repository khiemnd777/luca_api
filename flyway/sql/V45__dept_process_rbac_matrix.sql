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
  ('Công đoạn sản xuất - Xem', 'process.view'),
  ('Công đoạn sản xuất - Tạo', 'process.create'),
  ('Công đoạn sản xuất - Sửa', 'process.update'),
  ('Công đoạn sản xuất - Xoá', 'process.delete'),
  ('Công đoạn sản xuất - Tìm kiếm', 'process.search'),
	('Công đoạn sản xuất - Import', 'process.import'),
	('Công đoạn sản xuất - Export', 'process.export')
ON CONFLICT (permission_value)
DO UPDATE SET permission_name = EXCLUDED.permission_name;

-- ============================================
-- LINK ALL PERMISSIONS TO ADMIN ROLE
-- ============================================
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.permission_value IN (
  'process.view',
  'process.create',
  'process.update',
  'process.delete',
  'process.search',
	'process.import',
	'process.export'
)
WHERE r.role_name = 'admin'
ON CONFLICT DO NOTHING;
