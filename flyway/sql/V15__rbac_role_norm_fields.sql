CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION public.unaccent_immutable(text)
RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE RETURNS NULL ON NULL INPUT
AS $$ SELECT unaccent('unaccent'::regdictionary, $1) $$;

-- Users
ALTER TABLE roles
  ADD COLUMN IF NOT EXISTS display_name_norm  text GENERATED ALWAYS AS (unaccent_immutable(lower(display_name)))  STORED,
  ADD COLUMN IF NOT EXISTS role_name_norm     text GENERATED ALWAYS AS (unaccent_immutable(lower(role_name)))     STORED;

CREATE INDEX IF NOT EXISTS idx_rbac_role_display_name_trgm_norm ON roles USING gin (display_name_norm gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_rbac_role_role_name_trgm_norm    ON roles USING gin (role_name_norm gin_trgm_ops);
