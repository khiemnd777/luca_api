CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION public.unaccent_immutable(text)
RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE RETURNS NULL ON NULL INPUT
AS $$ SELECT unaccent('unaccent'::regdictionary, $1) $$;

-- Clinic
ALTER TABLE clinics
  ADD COLUMN IF NOT EXISTS name_norm  text GENERATED ALWAYS AS (unaccent_immutable(lower(name)))  STORED;

CREATE INDEX IF NOT EXISTS idx_clinic_name_trgm_norm  ON clinics USING gin (name_norm gin_trgm_ops);

-- Dentist
ALTER TABLE dentists
  ADD COLUMN IF NOT EXISTS name_norm  text GENERATED ALWAYS AS (unaccent_immutable(lower(name)))  STORED;

CREATE INDEX IF NOT EXISTS idx_dentist_name_trgm_norm  ON dentists USING gin (name_norm gin_trgm_ops);

-- Patient
ALTER TABLE patients
  ADD COLUMN IF NOT EXISTS name_norm  text GENERATED ALWAYS AS (unaccent_immutable(lower(name)))  STORED;

CREATE INDEX IF NOT EXISTS idx_patient_name_trgm_norm  ON patients USING gin (name_norm gin_trgm_ops);