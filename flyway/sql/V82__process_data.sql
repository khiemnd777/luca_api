UPDATE processes
SET department_id = 1
WHERE department_id IS NULL;

INSERT INTO processes (department_id, name, active, custom_fields, created_at, updated_at)
VALUES
	(1, 'Đai mẫu', TRUE, '{}'::jsonb, NOW(), NOW()),
	(1, 'Sáp', TRUE, '{}'::jsonb, NOW(), NOW()),
	(1, 'Cadcam', TRUE, '{}'::jsonb, NOW(), NOW()),
	(1, 'Sườn', TRUE, '{}'::jsonb, NOW(), NOW()),
	(1, 'Sứ', TRUE, '{}'::jsonb, NOW(), NOW()),
	(1, 'Tháo lắp', TRUE, '{}'::jsonb, NOW(), NOW()),
	(1, 'Cố định', TRUE, '{}'::jsonb, NOW(), NOW()),
	(1, 'Implant', TRUE, '{}'::jsonb, NOW(), NOW()),
	(1, 'Tháo lắp implant', TRUE, '{}'::jsonb, NOW(), NOW()),
	(1, 'Admin', TRUE, '{}'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;
