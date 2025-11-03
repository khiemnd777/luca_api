-- RBAC indexes: roles, permissions, user_roles, role_permissions

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'pg_trgm') THEN
    -- Sẽ tạo nếu chưa có; nếu không có quyền -> vẫn không ảnh hưởng migration (tiếp tục chạy phần còn lại)
    BEGIN
      CREATE EXTENSION IF NOT EXISTS pg_trgm;
    EXCEPTION
      WHEN insufficient_privilege THEN
        -- Không có quyền tạo extension -> bỏ qua
        NULL;
      WHEN duplicate_object THEN
        NULL;
    END;
  END IF;
END $$;

------------------------------------------------------------
-- 1) Unique case-insensitive cho roles.role_name
------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ux_roles_role_name_lower'
  ) THEN
    CREATE UNIQUE INDEX ux_roles_role_name_lower ON roles (LOWER(role_name));
  END IF;
END $$;

------------------------------------------------------------
-- 2) Unique case-insensitive cho permissions.permission_value
------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ux_permissions_value_lower'
  ) THEN
    CREATE UNIQUE INDEX ux_permissions_value_lower ON permissions (LOWER(permission_value));
  END IF;
END $$;

------------------------------------------------------------
-- 3) Btree giúp ORDER BY permission_value ASC mượt hơn (list)
------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ix_permissions_value_btree'
  ) THEN
    CREATE INDEX ix_permissions_value_btree ON permissions (permission_value ASC);
  END IF;
END $$;

------------------------------------------------------------
-- 4) M2M: user_roles
--    - Tối ưu ListByUser(user_id) và clear edges nhanh
--    - Đảm bảo unique (role_id, user_id) tránh gán trùng
------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ix_user_roles_user_role'
  ) THEN
    CREATE INDEX ix_user_roles_user_role ON user_roles (user_id, role_id);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ix_user_roles_role_user'
  ) THEN
    CREATE INDEX ix_user_roles_role_user ON user_roles (role_id, user_id);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ux_user_roles_role_user_unique'
  ) THEN
    CREATE UNIQUE INDEX ux_user_roles_role_user_unique ON user_roles (role_id, user_id);
  END IF;
END $$;

------------------------------------------------------------
-- 5) M2M: role_permissions
--    - Tối ưu PermissionIDsOfRole(role_id) và clear edges nhanh
--    - Đảm bảo unique (role_id, permission_id) tránh gán trùng
------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ix_role_permissions_role_perm'
  ) THEN
    CREATE INDEX ix_role_permissions_role_perm ON role_permissions (role_id, permission_id);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ix_role_permissions_perm_role'
  ) THEN
    CREATE INDEX ix_role_permissions_perm_role ON role_permissions (permission_id, role_id);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ux_role_permissions_role_perm_unique'
  ) THEN
    CREATE UNIQUE INDEX ux_role_permissions_role_perm_unique ON role_permissions (role_id, permission_id);
  END IF;
END $$;

------------------------------------------------------------
-- 6) (Optional) Trigram search indexes (display_name, brief, permission_name/value)
--    - CHỈ tạo nếu pg_trgm đã được cài đặt thành công.
------------------------------------------------------------
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm') THEN
    -- roles.display_name
    IF NOT EXISTS (
      SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ix_roles_display_name_trgm'
    ) THEN
      CREATE INDEX ix_roles_display_name_trgm ON roles USING gin (display_name gin_trgm_ops);
    END IF;

    -- roles.brief
    IF NOT EXISTS (
      SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ix_roles_brief_trgm'
    ) THEN
      CREATE INDEX ix_roles_brief_trgm ON roles USING gin (brief gin_trgm_ops);
    END IF;

    -- permissions.permission_name
    IF NOT EXISTS (
      SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ix_permissions_name_trgm'
    ) THEN
      CREATE INDEX ix_permissions_name_trgm ON permissions USING gin (permission_name gin_trgm_ops);
    END IF;

    -- permissions.permission_value (tìm kiếm mờ)
    IF NOT EXISTS (
      SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'ix_permissions_value_trgm'
    ) THEN
      CREATE INDEX ix_permissions_value_trgm ON permissions USING gin (permission_value gin_trgm_ops);
    END IF;
  END IF;
END $$;
