-- 1. Tạo permission "rbac.manage" nếu chưa tồn tại
INSERT INTO permissions (permission_name, permission_value)
SELECT 'RBAC Manage', 'rbac.manage'
WHERE NOT EXISTS (
  SELECT 1 FROM permissions WHERE permission_value = 'rbac.manage'
);

-- 2. Tạo role "admin" nếu chưa có
INSERT INTO roles (role_name)
SELECT 'admin'
WHERE NOT EXISTS (
  SELECT 1 FROM roles WHERE role_name = 'admin'
);

-- 3. Liên kết admin <-> rbac.manage
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.role_name = 'admin'
  AND p.permission_value = 'rbac.manage'
ON CONFLICT DO NOTHING;
