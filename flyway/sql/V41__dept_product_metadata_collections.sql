INSERT INTO collections (slug, name)
VALUES ('product', 'Sản phẩm')
ON CONFLICT (slug)
DO UPDATE SET name = EXCLUDED.name;
