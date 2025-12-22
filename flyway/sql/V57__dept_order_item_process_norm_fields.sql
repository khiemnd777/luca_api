CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION public.unaccent_immutable(text)
RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE RETURNS NULL ON NULL INPUT
AS $$ SELECT unaccent('unaccent'::regdictionary, $1) $$;

ALTER TABLE order_item_processes
  ADD COLUMN IF NOT EXISTS process_name_norm text GENERATED ALWAYS AS (unaccent_immutable(lower(process_name))) STORED,
  ADD COLUMN IF NOT EXISTS section_name_norm text GENERATED ALWAYS AS (unaccent_immutable(lower(section_name))) STORED;

CREATE INDEX IF NOT EXISTS idx_order_item_processes_process_name_trgm_norm  ON order_item_processes USING gin (process_name_norm gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_order_item_processes_section_name_trgm_norm  ON order_item_processes USING gin (section_name_norm gin_trgm_ops);
