CREATE INDEX IF NOT EXISTS ix_staff_user_id_deleted_at
  ON users(id, deleted_at);

CREATE INDEX IF NOT EXISTS ix_staff_user_id_not_deleted
  ON users(id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_staff_phone_not_deleted
  ON users(id, phone)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS ix_staff_email_not_deleted
  ON users(id, email)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_staff_staff_id_user_id 
  ON staffs(id, user_staff);