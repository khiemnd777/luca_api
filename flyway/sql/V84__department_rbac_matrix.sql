-- ======================================================
-- RBAC PERMISSIONS + CORPORATE ADMIN ROLE UPSERT SCRIPT
-- ======================================================

-- 1. Ensure role "corporate_admin" exists
INSERT INTO roles (role_name)
VALUES ('corporate_admin')
ON CONFLICT (role_name)
DO UPDATE SET role_name = EXCLUDED.role_name;

-- ============================================
-- PERMISSIONS UPSERT
-- ============================================
INSERT INTO permissions (permission_name, permission_value)
VALUES
  ('Chi nhánh - Xem', 'department.view'),
  ('Chi nhánh - Tạo', 'department.create'),
  ('Chi nhánh - Sửa', 'department.update'),
  ('Chi nhánh - Xoá', 'department.delete')
ON CONFLICT (permission_value)
DO UPDATE SET permission_name = EXCLUDED.permission_name;

-- ============================================
-- LINK ALL PERMISSIONS TO CORPORATE ADMIN ROLE
-- ============================================
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.permission_value IN (
  'department.view',
  'department.create',
  'department.update',
  'department.delete'
)
WHERE r.role_name = 'corporate_admin'
ON CONFLICT DO NOTHING;
