INSERT INTO collections (slug, name)
VALUES ('order-item', 'Đơn hàng phụ')
ON CONFLICT (slug)
DO UPDATE SET name = EXCLUDED.name;
