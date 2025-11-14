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
  ('Nhân sự - Tìm kiếm', 'staff.search'),
  ('Nha khoa - Tìm kiếm', 'clinic.search')
ON CONFLICT (permission_value)
DO UPDATE SET permission_name = EXCLUDED.permission_name;

-- ============================================
-- LINK ALL PERMISSIONS TO ADMIN ROLE
-- ============================================
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.permission_value IN (
  'staff.search',
  'clinic.search'
)
WHERE r.role_name = 'admin'
ON CONFLICT DO NOTHING;
