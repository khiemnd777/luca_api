INSERT INTO collections (slug, name)
VALUES ('material', 'Vật tư')
ON CONFLICT (slug)
DO UPDATE SET name = EXCLUDED.name;
