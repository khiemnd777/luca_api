-- A) Metadata cho schema động
CREATE TABLE IF NOT EXISTS collections (
  id SERIAL PRIMARY KEY,
  slug TEXT UNIQUE NOT NULL,     -- ví dụ: 'products', 'orders'
  name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS fields (
  id            SERIAL  PRIMARY KEY,
  collection_id INT     NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
  name          TEXT    NOT NULL,            -- key trong custom_fields
  label         TEXT    NOT NULL,
  type          TEXT    NOT NULL,            -- text, number, bool, date, select, multiselect, relation, json, richtext...
  required      BOOL    DEFAULT FALSE,
  "unique"      BOOL    DEFAULT FALSE,
  "table"       BOOL    DEFAULT FALSE,
  form          BOOL    DEFAULT FALSE,
  default_value JSONB,
  options       JSONB,                 -- { "choices":[...], "min":0, "max":999, ... }
  order_index   INT     DEFAULT 0,
  visibility    TEXT    DEFAULT 'public', -- public/admin/internal/...
  relation      JSONB                 -- { "target":"categories", "many":true, "fk":"category_id" }
);

-- Lưu ý các bảng có sử dụng cơ chế metadata driven, thì phải tạo một bảng custom_fields.
-- -- B) Ví dụ bảng nghiệp vụ có cột custom_fields
-- ALTER TABLE products
--   ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

-- -- Index GIN cho tìm kiếm động
-- CREATE INDEX IF NOT EXISTS idx_products_custom_gin ON products USING GIN (custom_fields);

-- -- Index biểu thức cho key hay lọc nhiều (ví dụ: color)
-- CREATE INDEX IF NOT EXISTS idx_products_cf_color ON products ((custom_fields->>'color'));
