CREATE UNIQUE INDEX IF NOT EXISTS raw_materials_dept_category_name_uq
ON raw_materials (department_id, category_id, name)
WHERE deleted_at IS NULL;

INSERT INTO raw_materials (department_id, category_id, category_name, name, created_at, updated_at)
SELECT
	1,
	c.id,
	c.name,
	v.name,
	NOW(),
	NOW()
FROM (VALUES
	('Tháo Lắp', 'Kim Loại Ti'),
	('Tháo Lắp', 'Kim Loại Cr-Co'),
	('Tháo Lắp', 'Cường Lực'),
	('Tháo Lắp', 'Nhựa Thường'),
	('Tháo Lắp', 'Nhựa Dẻo'),
	('Implant', 'Kim Loại'),
	('Implant', 'Không Kim Loại'),
	('Implant', 'Cường Lực'),
	('Implant', 'Nhựa Thường'),
	('Implant', 'PMMA'),
	('Implant', 'Khung Sườn'),
	('Implant', 'Khung Titan'),
	('Implant', 'Khung Cùi Titan'),
	('Implant', 'Bar Kim Loại'),
	('Implant', 'Trụ Abutment'),
	('Implant', 'Zirconia')
) AS v(category_name, name)
JOIN categories c
	ON c.name = v.category_name
	AND c.deleted_at IS NULL
ON CONFLICT DO NOTHING;
