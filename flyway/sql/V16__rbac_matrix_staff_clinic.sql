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
  ('Phân quyền', 'rbac.manage'),
  ('Nhân sự - Xem', 'staff.view'),
  ('Nhân sự - Tạo', 'staff.create'),
  ('Nhân sự - Sửa', 'staff.update'),
  ('Nhân sự - Xoá', 'staff.delete'),
  ('Nha khoa - Xem', 'clinic.view'),
  ('Nha khoa - Tạo', 'clinic.create'),
  ('Nha khoa - Sửa', 'clinic.update'),
  ('Nha khoa - Xoá', 'clinic.delete'),
  ('Settings - Xem', 'settings.view'),
  ('Settings - Sửa', 'settings.update')
ON CONFLICT (permission_value)
DO UPDATE SET permission_name = EXCLUDED.permission_name;

-- ============================================
-- LINK ALL PERMISSIONS TO ADMIN ROLE
-- ============================================
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.permission_value IN (
  'rbac.manage',
  'staff.view', 
  'staff.create', 
  'staff.update', 
  'staff.delete',
  'clinic.view', 
  'clinic.create', 
  'clinic.update', 
  'clinic.delete',
  'settings.view', 
  'settings.update'
)
WHERE r.role_name = 'admin'
ON CONFLICT DO NOTHING;
