INSERT INTO collections (slug, name)
VALUES ('order', 'Đơn hàng')
ON CONFLICT (slug)
DO UPDATE SET name = EXCLUDED.name;
