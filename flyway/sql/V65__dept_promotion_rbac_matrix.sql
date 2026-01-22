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
  ('Khuyến mãi - Xem', 'promotion.view'),
  ('Khuyến mãi - Tạo', 'promotion.create'),
  ('Khuyến mãi - Sửa', 'promotion.update'),
  ('Khuyến mãi - Xoá', 'promotion.delete'),
  ('Khuyến mãi - Tìm kiếm', 'promotion.search'),
	('Khuyến mãi - Import', 'promotion.import'),
	('Khuyến mãi - Export', 'promotion.export')
ON CONFLICT (permission_value)
DO UPDATE SET permission_name = EXCLUDED.permission_name;

-- ============================================
-- LINK ALL PERMISSIONS TO ADMIN ROLE
-- ============================================
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.permission_value IN (
  'promotion.view',
  'promotion.create',
  'promotion.update',
  'promotion.delete',
  'promotion.search',
	'promotion.import',
	'promotion.export'
)
WHERE r.role_name = 'admin'
ON CONFLICT DO NOTHING;
