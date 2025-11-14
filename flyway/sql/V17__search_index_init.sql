-- Extensions
CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- An immutable wrapper around unaccent using a fixed regdictionary
CREATE OR REPLACE FUNCTION f_unaccent(text)
RETURNS text
LANGUAGE sql
IMMUTABLE
PARALLEL SAFE
AS $$
  SELECT public.unaccent('public.unaccent'::regdictionary, $1);
$$;

-- Table
CREATE TABLE IF NOT EXISTS search_index (
  id            BIGSERIAL PRIMARY KEY,
  entity_type   TEXT NOT NULL,               -- 'user'|'department'|'section'|'product'|'order'|...
  entity_id     BIGINT NOT NULL,
  title         TEXT NOT NULL,
  subtitle      TEXT,
  keywords      TEXT,
  content       TEXT,
  attributes    JSONB NOT NULL DEFAULT '{}', -- per-type attributes (facets)
  org_id        BIGINT,
  owner_id      BIGINT,
  acl_hash      TEXT,                        -- hash của ACL (role/perm/scope) nếu cần
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

  -- FULL-TEXT (unaccent, weighted)
  tsv tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('simple', f_unaccent(coalesce(title,   ''))), 'A') ||
    setweight(to_tsvector('simple', f_unaccent(coalesce(subtitle,''))), 'B') ||
    setweight(to_tsvector('simple', f_unaccent(coalesce(keywords,''))), 'C') ||
    setweight(to_tsvector('simple', f_unaccent(coalesce(content, ''))), 'D')
  ) STORED,

  -- Trigram-friendly normalized text (for prefix/typo)
  norm TEXT GENERATED ALWAYS AS (
    f_unaccent(lower(
      coalesce(title,'') || ' ' ||
      coalesce(subtitle,'') || ' ' ||
      coalesce(keywords,'') || ' ' ||
      coalesce(content,'')
    ))
  ) STORED
);

-- Uniqueness per entity
CREATE UNIQUE INDEX IF NOT EXISTS ux_search_entity ON search_index(entity_type, entity_id);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_search_tsv       ON search_index USING GIN (tsv);
CREATE INDEX IF NOT EXISTS idx_search_norm_trgm ON search_index USING GIN (norm gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_search_attr_gin  ON search_index USING GIN (attributes);
CREATE INDEX IF NOT EXISTS idx_search_org       ON search_index (org_id);
CREATE INDEX IF NOT EXISTS idx_search_owner     ON search_index (owner_id);
