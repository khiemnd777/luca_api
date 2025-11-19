INSERT INTO collections (slug, name)
VALUES ('supplier', 'Nhà cung cấp')
ON CONFLICT (slug)
DO UPDATE SET name = EXCLUDED.name;
