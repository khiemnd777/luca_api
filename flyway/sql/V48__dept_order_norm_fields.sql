CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION public.unaccent_immutable(text)
RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE RETURNS NULL ON NULL INPUT
AS $$ SELECT unaccent('unaccent'::regdictionary, $1) $$;

ALTER TABLE orders
  ADD COLUMN IF NOT EXISTS code_norm text GENERATED ALWAYS AS (unaccent_immutable(lower(code))) STORED,
  ADD COLUMN IF NOT EXISTS customer_name_norm text GENERATED ALWAYS AS (unaccent_immutable(lower(customer_name))) STORED,
  ADD COLUMN IF NOT EXISTS code_latest_norm text GENERATED ALWAYS AS (unaccent_immutable(lower(code_latest))) STORED;

CREATE INDEX IF NOT EXISTS idx_order_code_trgm_norm  ON orders USING gin (code_norm gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_order_customer_name_trgm_norm  ON orders USING gin (customer_name_norm gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_order_code_latest_trgm_norm  ON orders USING gin (code_latest_norm gin_trgm_ops);
