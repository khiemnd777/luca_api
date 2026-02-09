CREATE UNIQUE INDEX IF NOT EXISTS restoration_types_category_name_uq
ON restoration_types (department_id, category_id, name)
WHERE deleted_at IS NULL;

INSERT INTO restoration_types (category_id, category_name, name, department_id, created_at, updated_at)
SELECT c.id, 'Implant', v.name, 1, NOW(), NOW()
FROM (VALUES
	('Implant', 'Cement'),
	('Implant', 'Cement Bắt Vít'),
	('Implant', 'Bắt Vít'),
	('Implant', 'Hàm Phủ')
) AS v(category_name, name)
JOIN categories c
	ON c.name = v.category_name
	AND c.deleted_at IS NULL
ON CONFLICT DO NOTHING;
