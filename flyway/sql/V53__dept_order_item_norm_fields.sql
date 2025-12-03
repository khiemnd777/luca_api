CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION public.unaccent_immutable(text)
RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE RETURNS NULL ON NULL INPUT
AS $$ SELECT unaccent('unaccent'::regdictionary, $1) $$;

ALTER TABLE order_items
  ADD COLUMN IF NOT EXISTS code_norm text GENERATED ALWAYS AS (unaccent_immutable(lower(code))) STORED;

CREATE INDEX IF NOT EXISTS idx_order_item_code_trgm_norm  ON order_items USING gin (code_norm gin_trgm_ops);
