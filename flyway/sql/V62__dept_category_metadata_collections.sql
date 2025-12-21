INSERT INTO collections (slug, name)
VALUES ('category', 'Danh mục')
ON CONFLICT (slug)
DO UPDATE SET name = EXCLUDED.name;
