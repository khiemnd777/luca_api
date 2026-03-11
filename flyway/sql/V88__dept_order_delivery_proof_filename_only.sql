UPDATE order_delivery_proofs
SET image_url = regexp_replace(image_url, '^.*/', '')
WHERE image_url LIKE '%/%';
