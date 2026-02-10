CREATE UNIQUE INDEX IF NOT EXISTS brand_names_dept_category_name_uq
ON brand_names (department_id, category_id, name)
WHERE deleted_at IS NULL;

-- Seed base brand names (additive only)
INSERT INTO brand_names (department_id, category_id, category_name, name, created_at, updated_at)
SELECT
	1,
	c.id,
	c.name,
	v.name,
	NOW(),
	NOW()
FROM (VALUES
	('Cố Định', 'Kerox'),
	('Cố Định', 'Cercon'),
	('Cố Định', 'Lava'),
	('Cố Định', 'Vita'),
	('Tháo Lắp', 'Răng Nhật'),
	('Tháo Lắp', 'Răng Mỹ'),
	('Tháo Lắp', 'Răng Composite'),
	('Tháo Lắp', 'Răng Enigmalife'),
	('Implant', 'Ni-Cr'),
	('Implant', 'Ti'),
	('Implant', 'Cr-Co'),
	('Implant', 'Kerox'),
	('Implant', 'Lava'),
	('Implant', 'Cercon'),
	('Implant', 'Vita'),
	('Implant', 'Sứ Nano'),
	('Implant', 'Răng Mỹ'),
	('Implant', 'Sứ Ni-Cr'),
	('Implant', 'Sứ Cr-Co'),
	('Implant', 'Sứ Titan'),
	('Implant', 'Răng Composite'),
	('Implant', 'Mắc Cài Rhein 83'),
	('Implant', 'Titan'),
	('Implant', 'Hàn Quốc'),
	('Implant', 'Châu Âu'),
	('Implant', 'Bar Thuỵ Sỹ')
) AS v(category_name, name)
JOIN categories c
	ON c.name = v.category_name
	AND c.deleted_at IS NULL
ON CONFLICT DO NOTHING;
