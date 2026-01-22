package promotionmodel

import (
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotioncode"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotioncondition"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotionscope"
)

type PromotionScopeType = promotionscope.ScopeType

const (
	PromotionScopeAll      PromotionScopeType = promotionscope.ScopeTypeALL
	PromotionScopeUser     PromotionScopeType = promotionscope.ScopeTypeUSER
	PromotionScopeSeller   PromotionScopeType = promotionscope.ScopeTypeSELLER
	PromotionScopeCategory PromotionScopeType = promotionscope.ScopeTypeCATEGORY
	PromotionScopeProduct  PromotionScopeType = promotionscope.ScopeTypePRODUCT
)

type PromotionConditionType = promotioncondition.ConditionType

const (
	PromotionConditionOrderIsRemake    PromotionConditionType = promotioncondition.ConditionTypeORDER_IS_REMAKE
	PromotionConditionRemakeCountLTE   PromotionConditionType = promotioncondition.ConditionTypeREMAKE_COUNT_LTE
	PromotionConditionRemakeWithinDays PromotionConditionType = promotioncondition.ConditionTypeREMAKE_WITHIN_DAYS
	PromotionConditionRemakeReason     PromotionConditionType = promotioncondition.ConditionTypeREMAKE_REASON
)

type PromotionDiscountType = promotioncode.DiscountType

const (
	PromotionDiscountFixed        PromotionDiscountType = promotioncode.DiscountTypeFixed
	PromotionDiscountPercent      PromotionDiscountType = promotioncode.DiscountTypePercent
	PromotionDiscountFreeShipping PromotionDiscountType = promotioncode.DiscountTypeFreeShipping
)

func ValidateScopeInput(scopeType PromotionScopeType) error {
	return promotionscope.ScopeTypeValidator(promotionscope.ScopeType(scopeType))
}

func ValidateConditionInput(conditionType PromotionConditionType) error {
	return promotioncondition.ConditionTypeValidator(promotioncondition.ConditionType(conditionType))
}
