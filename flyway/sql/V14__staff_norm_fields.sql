CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION public.unaccent_immutable(text)
RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE RETURNS NULL ON NULL INPUT
AS $$ SELECT unaccent('unaccent'::regdictionary, $1) $$;

-- Users
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS name_norm  text GENERATED ALWAYS AS (unaccent_immutable(lower(name)))  STORED,
  ADD COLUMN IF NOT EXISTS phone_norm  text GENERATED ALWAYS AS (unaccent_immutable(lower(phone)))  STORED,
  ADD COLUMN IF NOT EXISTS email_norm  text GENERATED ALWAYS AS (unaccent_immutable(lower(email)))  STORED;

CREATE INDEX IF NOT EXISTS idx_user_name_trgm_norm  ON users USING gin (name_norm gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_user_phone_trgm_norm  ON users USING gin (phone_norm gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_user_email_trgm_norm  ON users USING gin (email_norm gin_trgm_ops);
