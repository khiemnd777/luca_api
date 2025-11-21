INSERT INTO collections (slug, name)
VALUES ('process', 'Công đoạn sản xuất')
ON CONFLICT (slug)
DO UPDATE SET name = EXCLUDED.name;
