CREATE UNIQUE INDEX IF NOT EXISTS idx_promotion_codes_code
  ON promotion_codes (code);

CREATE INDEX IF NOT EXISTS idx_promotion_usages_promo_user
  ON promotion_usages (promo_code_id, user_id);

CREATE INDEX IF NOT EXISTS idx_promotion_usages_promo
  ON promotion_usages (promo_code_id);

CREATE INDEX IF NOT EXISTS idx_order_promotions_order
  ON order_promotions (order_id);

CREATE INDEX IF NOT EXISTS idx_order_promotions_promo_code
  ON order_promotions (promo_code);

CREATE INDEX IF NOT EXISTS idx_promotion_codes_created_at
  ON promotion_codes (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_promotion_codes_active_created
  ON promotion_codes (is_active, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_promotion_codes_active_window
  ON promotion_codes (is_active, start_at, end_at);
