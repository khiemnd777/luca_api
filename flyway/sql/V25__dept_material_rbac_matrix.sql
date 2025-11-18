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
  ('Vật tư - Xem', 'material.view'),
  ('Vật tư - Tạo', 'material.create'),
  ('Vật tư - Sửa', 'material.update'),
  ('Vật tư - Xoá', 'material.delete'),
  ('Vật tư - Tìm kiếm', 'material.search'),
	('Vật tư - Import', 'material.import'),
	('Vật tư - Export', 'material.export')
ON CONFLICT (permission_value)
DO UPDATE SET permission_name = EXCLUDED.permission_name;

-- ============================================
-- LINK ALL PERMISSIONS TO ADMIN ROLE
-- ============================================
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.permission_value IN (
  'material.view',
  'material.create',
  'material.update',
  'material.delete',
  'material.search',
	'material.import',
	'material.export'
)
WHERE r.role_name = 'admin'
ON CONFLICT DO NOTHING;
