CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION public.unaccent_immutable(text)
RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE RETURNS NULL ON NULL INPUT
AS $$ SELECT unaccent('unaccent'::regdictionary, $1) $$;

ALTER TABLE customers
  ADD COLUMN IF NOT EXISTS code_norm text GENERATED ALWAYS AS (unaccent_immutable(lower(code))) STORED,
	ADD COLUMN IF NOT EXISTS name_norm text GENERATED ALWAYS AS (unaccent_immutable(lower(name))) STORED;

CREATE INDEX IF NOT EXISTS idx_customer_code_trgm_norm  ON customers USING gin (code_norm gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_customer_name_trgm_norm  ON customers USING gin (name_norm gin_trgm_ops);
