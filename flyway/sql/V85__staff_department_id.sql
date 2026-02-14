ALTER TABLE staffs
  ADD COLUMN IF NOT EXISTS department_id INT;

CREATE INDEX IF NOT EXISTS ix_staff_department_id
  ON staffs(department_id);

CREATE INDEX IF NOT EXISTS ix_staff_department_user_id
  ON staffs(department_id, user_staff);
