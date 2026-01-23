package engine

import (
	"time"

	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

type OrderContext struct {
	TotalPrice     float64
	IsRemake       bool
	RemakeCount    int
	RemakeReason   string
	OriginalTime   time.Time
	ProductIDs     []int
	ShippingAmount float64
	SellerID       int
}

type PromotionApplyResult struct {
	DiscountAmount    float64
	FinalPrice        float64
	AppliedConditions []string
	PromoCode         string
	Promotion         *generated.PromotionCode
}
