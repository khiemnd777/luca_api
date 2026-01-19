package model

import (
	"time"

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
	ID                int                       `json:"id,omitempty"`
	Code              string                    `json:"code,omitempty"`
	Name              *string                   `json:"name,omitempty"`
	DiscountType      promotioncode.DiscountType `json:"discount_type,omitempty"`
	DiscountValue     int                       `json:"discount_value,omitempty"`
	MaxDiscountAmount *int                      `json:"max_discount_amount,omitempty"`
	MinOrderValue     *int                      `json:"min_order_value,omitempty"`
	TotalUsageLimit   *int                      `json:"total_usage_limit,omitempty"`
	UsagePerUser      *int                      `json:"usage_per_user,omitempty"`
	StartAt           time.Time                 `json:"start_at,omitempty"`
	EndAt             time.Time                 `json:"end_at,omitempty"`
	IsActive          bool                      `json:"is_active,omitempty"`
	CreatedAt         time.Time                 `json:"created_at,omitempty"`
	UpdatedAt         time.Time                 `json:"updated_at,omitempty"`
}

type CreatePromotionInput struct {
	Code              string
	DiscountType      promotioncode.DiscountType
	DiscountValue     int
	MaxDiscountAmount *int
	MinOrderValue     *int
	TotalUsageLimit   *int
	UsagePerUser      *int
	StartAt           *time.Time
	EndAt             *time.Time
	IsActive          bool
}

type UpdatePromotionInput struct {
	DiscountType      promotioncode.DiscountType
	DiscountValue     int
	MaxDiscountAmount *int
	MinOrderValue     *int
	TotalUsageLimit   *int
	UsagePerUser      *int
	StartAt           *time.Time
	EndAt             *time.Time
	IsActive          *bool
}
