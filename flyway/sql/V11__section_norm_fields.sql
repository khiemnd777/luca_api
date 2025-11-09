CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION public.unaccent_immutable(text)
RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE RETURNS NULL ON NULL INPUT
AS $$ SELECT unaccent('unaccent'::regdictionary, $1) $$;

ALTER TABLE sections
  ADD COLUMN IF NOT EXISTS name_norm  text GENERATED ALWAYS AS (unaccent_immutable(lower(name)))  STORED;

CREATE INDEX IF NOT EXISTS idx_sections_name_trgm_norm  ON sections USING gin (name_norm gin_trgm_ops);