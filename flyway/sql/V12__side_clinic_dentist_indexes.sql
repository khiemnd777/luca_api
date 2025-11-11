CREATE INDEX IF NOT EXISTS ix_clinic_id_not_deleted
  ON clinics(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_dentist_id_not_deleted
  ON dentists(id)
  WHERE deleted_at IS NULL;