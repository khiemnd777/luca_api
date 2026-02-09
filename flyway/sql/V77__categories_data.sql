-- Unique partial indexes to prevent duplicates by level
CREATE UNIQUE INDEX IF NOT EXISTS categories_lv1_name_uq
ON categories (department_id, name)
WHERE level = 1 AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS categories_lv2_parent_name_uq
ON categories (department_id, parent_id, name)
WHERE level = 2 AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS categories_lv3_parent_name_uq
ON categories (department_id, parent_id, name)
WHERE level = 3 AND deleted_at IS NULL;

-- Seed base categories (additive only)
INSERT INTO categories (name, level, active, custom_fields, department_id, created_at, updated_at)
VALUES
	('Cố Định', 1, TRUE, '{}'::jsonb, 1, NOW(), NOW()),
	('Tháo Lắp', 1, TRUE, '{}'::jsonb, 1, NOW(), NOW()),
	('Implant', 1, TRUE, '{}'::jsonb, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO categories (
	name, level, parent_id,
	category_id_lv1, category_name_lv1,
	active, custom_fields, department_id, created_at, updated_at
)
SELECT
	v.lv2_name,
	2,
	c1.id,
	c1.id,
	c1.name,
	TRUE,
	'{}'::jsonb,
	1,
	NOW(),
	NOW()
FROM (VALUES
	('Cố Định', 'Không Kim Loại'),
	('Cố Định', 'Kim Loại'),
	('Tháo Lắp', 'Hàm Khung'),
	('Tháo Lắp', 'Hàm Khung Liên Kết'),
	('Tháo Lắp', 'Bán Hàm'),
	('Tháo Lắp', 'Toàn Hàm'),
	('Tháo Lắp', 'Phụ Kiện Tháo Lắp'),
	('Implant', 'Sứ Trên Implant'),
	('Implant', 'Hàm Lai'),
	('Implant', 'Hàm OT Bridge'),
	('Implant', 'Hàm Bar'),
	('Implant', 'Sản Phẩm Khác')
) AS v(lv1_name, lv2_name)
JOIN categories c1
	ON c1.level = 1
	AND c1.name = v.lv1_name
	AND c1.deleted_at IS NULL
ON CONFLICT DO NOTHING;

INSERT INTO categories (
	name, level, parent_id,
	category_id_lv1, category_name_lv1,
	category_id_lv2, category_name_lv2,
	active, custom_fields, department_id, created_at, updated_at
)
SELECT
	v.lv3_name,
	3,
	c2.id,
	c1.id,
	c1.name,
	c2.id,
	c2.name,
	TRUE,
	'{}'::jsonb,
	1,
	NOW(),
	NOW()
FROM (VALUES
	('Cố Định', 'Không Kim Loại', 'Full'),
	('Cố Định', 'Không Kim Loại', 'Veneer'),
	('Cố Định', 'Không Kim Loại', 'Đắp Sứ'),
	('Cố Định', 'Không Kim Loại', 'Làm Sườn'),
	('Cố Định', 'Không Kim Loại', 'Onlay'),
	('Cố Định', 'Không Kim Loại', 'Inlay'),
	('Cố Định', 'Không Kim Loại', 'Cùi Giả'),
	('Cố Định', 'Không Kim Loại', 'Răng Tạm'),
	('Cố Định', 'Không Kim Loại', 'Sứ Ép'),
	('Cố Định', 'Kim Loại', 'Full'),
	('Cố Định', 'Kim Loại', 'Veneer'),
	('Cố Định', 'Kim Loại', 'Đắp Sứ'),
	('Cố Định', 'Kim Loại', 'Làm Sườn'),
	('Cố Định', 'Kim Loại', 'Onlay'),
	('Cố Định', 'Kim Loại', 'Inlay'),
	('Cố Định', 'Kim Loại', 'Cùi Giả'),
	('Cố Định', 'Kim Loại', 'Mắc Cài')
) AS v(lv1_name, lv2_name, lv3_name)
JOIN categories c1
	ON c1.level = 1
	AND c1.name = v.lv1_name
	AND c1.deleted_at IS NULL
JOIN categories c2
	ON c2.level = 2
	AND c2.name = v.lv2_name
	AND c2.parent_id = c1.id
	AND c2.deleted_at IS NULL
ON CONFLICT DO NOTHING;
