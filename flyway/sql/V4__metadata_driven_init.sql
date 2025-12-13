-- A) Metadata cho schema động
CREATE TABLE IF NOT EXISTS collections (
  id          SERIAL      PRIMARY KEY,
  slug        TEXT        UNIQUE  NOT NULL,     -- ví dụ: 'products', 'orders', Uuid (when integration = TRUE)
  show_if     JSONB       NULL,
  integration BOOL        DEFAULT FALSE,        -- integrate directly with specific business e.g. category, product,...
  "group"     TEXT        NULL,                 -- group of integration e.g. category, product,...
  name        TEXT        NOT NULL,
  deleted_at  TIMESTAMPTZ NULL
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
  tag           TEXT    NULL,
  form          BOOL    DEFAULT FALSE,
  search        BOOL    DEFAULT FALSE,
  default_value JSONB,
  options       JSONB,                    -- { "choices":[...], "min":0, "max":999, ... }
  order_index   INT     DEFAULT 0,
  visibility    TEXT    DEFAULT 'public', -- public/admin/internal/...
  relation      JSONB                     -- { "target":"categories", "many":true, "fk":"category_id" }
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_collections_integration_group_slug
    ON collections (integration, "group", slug)
    WHERE deleted_at IS NULL;;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_collections_slug_trgm
    ON collections
    USING gin (slug gin_trgm_ops)
    WHERE deleted_at IS NULL;;

CREATE INDEX IF NOT EXISTS idx_collections_name_trgm
    ON collections
    USING gin (name gin_trgm_ops)
    WHERE deleted_at IS NULL;;

CREATE INDEX IF NOT EXISTS idx_fields_collection_id
    ON fields (collection_id);


-- Lưu ý các bảng có sử dụng cơ chế metadata driven, thì phải tạo một bảng custom_fields.
-- -- B) Ví dụ bảng nghiệp vụ có cột custom_fields
-- ALTER TABLE products
--   ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}'::jsonb;

-- -- Index GIN cho tìm kiếm động
-- CREATE INDEX IF NOT EXISTS idx_products_custom_gin ON products USING GIN (custom_fields);

-- -- Index biểu thức cho key hay lọc nhiều (ví dụ: color)
-- CREATE INDEX IF NOT EXISTS idx_products_cf_color ON products ((custom_fields->>'color'));
