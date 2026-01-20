package model

import (
	"encoding/json"
	"time"

	promotionmodel "github.com/khiemnd777/andy_api/modules/main/features/promotion/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotioncode"
)

type PromotionSnapshot struct {
	PromoCode         string    `json:"promo_code"`
	DiscountType      string    `json:"discount_type"`
	DiscountValue     int       `json:"discount_value"`
	DiscountAmount    float64   `json:"discount_amount"`
	IsRemake          bool      `json:"is_remake"`
	RemakeCount       int       `json:"remake_count"`
	AppliedConditions []string  `json:"applied_conditions"`
	AppliedAt         time.Time `json:"applied_at"`
}

type PromotionCodeDTO struct {
	ID                int                        `json:"id,omitempty"`
	Code              string                     `json:"code,omitempty"`
	Name              *string                    `json:"name,omitempty"`
	DiscountType      promotioncode.DiscountType `json:"discount_type,omitempty"`
	DiscountValue     int                        `json:"discount_value,omitempty"`
	MaxDiscountAmount *int                       `json:"max_discount_amount,omitempty"`
	MinOrderValue     *int                       `json:"min_order_value,omitempty"`
	TotalUsageLimit   *int                       `json:"total_usage_limit,omitempty"`
	UsagePerUser      *int                       `json:"usage_per_user,omitempty"`
	StartAt           time.Time                  `json:"start_at,omitempty"`
	EndAt             time.Time                  `json:"end_at,omitempty"`
	IsActive          bool                       `json:"is_active,omitempty"`
	Scopes            []PromotionScopeInput      `json:"scopes,omitempty"`
	Conditions        []PromotionConditionInput  `json:"conditions,omitempty"`
	CreatedAt         time.Time                  `json:"created_at,omitempty"`
	UpdatedAt         time.Time                  `json:"updated_at,omitempty"`
}

type PromotionScopeInput struct {
	ScopeType  promotionmodel.PromotionScopeType `json:"scope_type"`
	ScopeValue json.RawMessage                   `json:"scope_value"`
}

type PromotionConditionInput struct {
	ConditionType  promotionmodel.PromotionConditionType `json:"condition_type"`
	ConditionValue json.RawMessage                       `json:"condition_value"`
}

/*
e.g. payload:

	{
	  "code": "REMAKE10",
	  "discount_type": "percent",
	  "discount_value": 10,
	  "is_active": true,
	  "scopes": [
	    {
	      "scope_type": "ALL",
	      "scope_value": null
	    }
	  ],
	  "conditions": [
	    {
	      "condition_type": "ORDER_IS_REMAKE",
	      "condition_value": null
	    },
	    {
	      "condition_type": "REMAKE_COUNT_LTE",
	      "condition_value": 2
	    }
	  ]
	}
*/
type CreatePromotionInput struct {
	Code              string                     `json:"code"`
	DiscountType      promotioncode.DiscountType `json:"discount_type,omitempty"`
	DiscountValue     int                        `json:"discount_value,omitempty"`
	MaxDiscountAmount *int                       `json:"max_discount_amount,omitempty"`
	MinOrderValue     *int                       `json:"min_order_value,omitempty"`
	TotalUsageLimit   *int                       `json:"total_usage_limit,omitempty"`
	UsagePerUser      *int                       `json:"usage_per_user,omitempty"`
	StartAt           time.Time                  `json:"start_at,omitempty"`
	EndAt             time.Time                  `json:"end_at,omitempty"`
	IsActive          bool                       `json:"is_active"`
	Scopes            []PromotionScopeInput      `json:"scopes"`
	Conditions        []PromotionConditionInput  `json:"conditions"`
}

type UpdatePromotionInput struct {
	DiscountType      promotioncode.DiscountType `json:"discount_type,omitempty"`
	DiscountValue     int                        `json:"discount_value,omitempty"`
	MaxDiscountAmount *int                       `json:"max_discount_amount,omitempty"`
	MinOrderValue     *int                       `json:"min_order_value,omitempty"`
	TotalUsageLimit   *int                       `json:"total_usage_limit,omitempty"`
	UsagePerUser      *int                       `json:"usage_per_user,omitempty"`
	StartAt           time.Time                  `json:"start_at,omitempty"`
	EndAt             time.Time                  `json:"end_at,omitempty"`
	IsActive          bool                       `json:"is_active"`
	Scopes            []PromotionScopeInput      `json:"scopes"`
	Conditions        []PromotionConditionInput  `json:"conditions"`
}
