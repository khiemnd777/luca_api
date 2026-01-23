CREATE UNIQUE INDEX IF NOT EXISTS idx_promotion_codes_code
  ON promotion_codes (code);

CREATE INDEX IF NOT EXISTS idx_promotion_usages_promo_user
  ON promotion_usages (promo_code_id, user_id);

CREATE INDEX IF NOT EXISTS idx_promotion_usages_promo
  ON promotion_usages (promo_code_id);

CREATE INDEX IF NOT EXISTS idx_promotion_usages_promo_for_order
  ON promotion_usages (order_id, promo_code_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_promotion_usage_order_promo_discount
  ON promotion_usages (order_id, promo_code_id, discount_amount);
  
CREATE INDEX IF NOT EXISTS idx_promotion_codes_created_at
  ON promotion_codes (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_promotion_codes_active_created
  ON promotion_codes (is_active, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_promotion_codes_active_window
  ON promotion_codes (is_active, start_at, end_at);

