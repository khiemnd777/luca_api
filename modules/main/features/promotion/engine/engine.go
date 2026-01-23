package engine

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
)

type Engine struct {
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewEngine(deps *module.ModuleDeps[config.ModuleConfig]) *Engine {
	return &Engine{deps: deps}
}

/*
	e.g.:

discount, applied, err := engine.NewEngine(s.deps).Apply(

	ctx,
	promo,
	userID,
	orderCtx,
	time.Now(),

)
*/
func (e *Engine) Apply(
	ctx context.Context,
	promo *generated.PromotionCode,
	userID int,
	orderCtx OrderContext,
	now time.Time,
) (*PromotionApplyResult, error) {

	// ===== SCOPE =====
	scopeMatched, err := e.matchScopes(ctx, promo, userID, orderCtx)
	if err != nil {
		return nil, err
	}
	if !scopeMatched {
		return nil, PromotionApplyError{Reason: "promotion_scope_not_matched"}
	}

	// ===== CONDITIONS =====
	appliedConditions, err := e.matchConditions(promo, orderCtx, now)
	if err != nil {
		return nil, err
	}

	// ===== DISCOUNT =====
	discountAmount, err := e.calculateDiscount(promo, orderCtx)
	if err != nil {
		return nil, err
	}

	finalPrice := orderCtx.TotalPrice - discountAmount
	if finalPrice < 0 {
		finalPrice = 0
	}

	return &PromotionApplyResult{
		DiscountAmount:    discountAmount,
		FinalPrice:        finalPrice,
		AppliedConditions: appliedConditions,
		PromoCode:         promo.Code,
		Promotion:         promo,
	}, nil
}
