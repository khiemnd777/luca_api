-- ------------------------------------------------
-- table profiles
-- ------------------------------------------------
CREATE TABLE IF NOT EXISTS import_field_profiles (
  id SERIAL PRIMARY KEY,

  -- entity / scope áp dụng profile này
  -- ví dụ: 'clinics', 'dentists', 'orders', 'products'
  scope TEXT NOT NULL,

  -- mã profile ngắn gọn để code dùng
  code TEXT NOT NULL,    -- ví dụ: 'default', 'vietnam-template-2025'

  -- tên hiển thị cho user
  name TEXT NOT NULL,    -- ví dụ: 'Default Clinic Import', 'Template Excel Bộ Y Tế'

  description TEXT,
  is_default BOOL NOT NULL DEFAULT FALSE,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_import_field_profiles_scope_code
  ON import_field_profiles (scope, code);

-- ------------------------------------------------
-- table field mappings
-- ------------------------------------------------
CREATE TABLE IF NOT EXISTS import_field_mappings (
  id SERIAL PRIMARY KEY,

  -- profile chứa mapping này
  profile_id INT NOT NULL REFERENCES import_field_profiles(id) ON DELETE CASCADE,

  -- INTERNAL FIELD (OWN FIELD)
  -- -------------------------
  -- core | metadata | external
  internal_kind TEXT NOT NULL,

  -- đường dẫn / key tới field:
  --   core:      'name', 'phone_number', 'address'
  --   metadata:  'tax_code', 'clinic_code' (key trong custom_fields)
  --   external:  'full_name_raw', 'import_note', ...
  internal_path  TEXT NOT NULL,

  -- label hiển thị cho field nội bộ (UI mapping)
  internal_label TEXT NOT NULL,

  -- chỉ dùng khi internal_kind = 'metadata'
  -- lưu text, KHÔNG FK sang collections/fields
  metadata_collection_slug TEXT,  -- ví dụ: 'clinics'
  metadata_field_name      TEXT,  -- ví dụ: 'tax_code'

  -- type gợi ý (để validate / transform khi import)
  data_type TEXT,  -- ví dụ: 'text', 'number', 'date', ... (có thể reuse type bên fields)

  -- EXCEL FIELD
  -- -------------------------
  excel_header TEXT, -- text header trong Excel: 'Clinic Name', 'Tax Code', ...
  excel_column INT,  -- index cột 1-based: A=1, B=2, ...

  required BOOL NOT NULL DEFAULT FALSE,
  "unique" BOOL NOT NULL DEFAULT FALSE,

  -- optional: hint transform (sau này xài cũng được)
  -- ví dụ: 'trim|upper', 'date:dd/MM/yyyy', ...
  transform_hint TEXT
);

-- 1 internal field chỉ nên có 1 mapping trong 1 profile
CREATE UNIQUE INDEX IF NOT EXISTS ux_import_field_mappings_profile_internal
  ON import_field_mappings (profile_id, internal_kind, internal_path);

