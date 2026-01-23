package engine

const (
	// generic
	ReasonPromotionNotFound   = "promotion_not_found"
	ReasonPromotionInactive   = "promotion_inactive"
	ReasonPromotionNotStarted = "promotion_not_started"
	ReasonPromotionExpired    = "promotion_expired"

	// scope
	ReasonPromotionScopeNotMatched = "promotion_scope_not_matched"

	// usage
	ReasonPromotionTotalUsageLimitReached = "promotion_total_usage_limit_reached"
	ReasonPromotionUserUsageLimitReached  = "promotion_user_usage_limit_reached"

	// order required / input
	ReasonPromoCodeRequired = "promo_code_required"
	ReasonOrderRequired     = "order_required"

	// ===== Conditions =====
	ReasonConditionOrderIsRemakeNotMet    = "condition_order_is_remake_not_met"
	ReasonConditionRemakeCountLTENotMet   = "condition_remake_count_lte_not_met"
	ReasonConditionRemakeWithinDaysNotMet = "condition_remake_within_days_not_met"
	ReasonConditionRemakeReasonNotMet     = "condition_remake_reason_not_met"

	// ===== Discount =====
	ReasonMinOrderValueNotMet = "min_order_value_not_met"
)
