-- 1) Tạo department "New Company" nếu chưa có (dựa trên slug)
INSERT INTO departments (
  name, slug, active, logo, address, phone_number, parent_id, created_at, updated_at
)
SELECT
  'New Company',
  'new_company',
  TRUE,
  'https://api.dicebear.com/9.x/initials/svg?seed=Company',
  NULL,
  NULL,
  NULL,
  NOW(),
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM departments WHERE slug = 'new_company'
);

-- 2) Gắn TẤT CẢ user có role_name='admin' vào department 'new_company'
INSERT INTO department_members (user_id, department_id, created_at)
SELECT
  u.id AS user_id,
  d.id AS department_id,
  NOW() AS created_at
FROM departments d
JOIN roles r         ON r.role_name = 'admin'
JOIN user_roles ur   ON ur.role_id  = r.id
JOIN users u         ON u.id        = ur.user_id
WHERE d.slug = 'new_company'
  AND NOT EXISTS (
    SELECT 1
    FROM department_members dm
    WHERE dm.user_id = u.id
      AND dm.department_id = d.id
  );
