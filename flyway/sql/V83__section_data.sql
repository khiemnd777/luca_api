UPDATE sections
SET department_id = 1
WHERE department_id IS NULL;

INSERT INTO sections (department_id, name, color, active, custom_fields, created_at, updated_at)
VALUES
	(1, 'Cố Định', '#2bbcda', TRUE, '{}'::jsonb, NOW(), NOW()),
	(1, 'Tháo Lắp', '#4b65e7', TRUE, '{}'::jsonb, NOW(), NOW()),
	(1, 'Implant', '#07e454', TRUE, '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;
