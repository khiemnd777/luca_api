INSERT INTO collections (slug, name)
VALUES ('customer', 'Khách hàng')
ON CONFLICT (slug)
DO UPDATE SET name = EXCLUDED.name;
