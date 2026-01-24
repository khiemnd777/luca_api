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
	guard PromotionGuard,
	userID *int,
	orderID int64,
	orderCtx OrderContext,
	now time.Time,
) (*PromotionApplyResult, error) {

	// ===== PRE-CHECK (repo-backed via guard) =====
	if err := guard.EnsureValidPromo(ctx, promo, now, orderID); err != nil {
		return nil, err
	}

	if err := guard.CheckUsage(ctx, promo, userID, orderID); err != nil {
		return nil, err
	}

	// ===== SCOPE =====
	scopeMatched, err := e.matchScopes(ctx, promo, userID, orderCtx)
	if err != nil {
		return nil, err
	}
	if !scopeMatched {
		return nil, PromotionApplyError{Reason: ReasonPromotionScopeNotMatched}
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
