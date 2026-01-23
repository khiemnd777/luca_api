package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/promotion/engine"
	promotionmodel "github.com/khiemnd777/andy_api/modules/main/features/promotion/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/product"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type promotionApplyResult struct {
	DiscountAmount    float64
	AppliedConditions []string
	Promo             *generated.PromotionCode
}

func (r *orderRepository) applyPromotion(
	ctx context.Context,
	userID int,
	order *model.OrderDTO,
) (*engine.PromotionApplyResult, *model.PromotionSnapshot) {

	if order == nil || order.PromotionCode == nil || order.PromotionCodeID == nil {
		logger.Debug(
			"skip promotion apply: missing promotion info",
		)
		return nil, nil
	}

	logger.Debug(
		"start apply promotion",
		"order_id", order.ID,
		"user_id", userID,
		"promo_code", *order.PromotionCode,
	)

	promo, err := r.promotionRepo.GetByCode(ctx, *order.PromotionCode)
	if err != nil {
		logger.Warn(
			"promotion not found or failed to load",
			"order_id", order.ID,
			"promo_code", *order.PromotionCode,
			"err", err,
		)
		return nil, nil
	}

	orderCtx := r.promoctxbuilder.BuildFromOrderDTO(order)

	logger.Debug(
		"promotion order context built",
		"order_id", order.ID,
		"total_price", orderCtx.TotalPrice,
		"is_remake", orderCtx.IsRemake,
		"remake_count", orderCtx.RemakeCount,
		"product_ids", orderCtx.ProductIDs,
		"seller_id", orderCtx.SellerID,
		"shipping_amount", orderCtx.ShippingAmount,
	)

	result, err := r.promoengine.Apply(
		ctx,
		promo,
		userID,
		orderCtx,
		time.Now(),
	)
	if err != nil {
		logger.Info(
			"promotion apply failed",
			"order_id", order.ID,
			"promo_code", promo.Code,
			"user_id", userID,
			"reason", err,
		)
		return nil, nil
	}

	logger.Info(
		"promotion applied successfully",
		"order_id", order.ID,
		"promo_code", result.PromoCode,
		"discount_amount", result.DiscountAmount,
		"applied_conditions", result.AppliedConditions,
	)

	snapshot := &model.PromotionSnapshot{
		PromoCode:         result.PromoCode,
		DiscountType:      string(result.Promotion.DiscountType),
		DiscountValue:     result.Promotion.DiscountValue,
		DiscountAmount:    result.DiscountAmount,
		IsRemake:          orderCtx.IsRemake,
		RemakeCount:       orderCtx.RemakeCount,
		AppliedConditions: result.AppliedConditions,
		AppliedAt:         time.Now(),
	}

	return result, snapshot
}

func (r *orderRepository) buildPromotionSnapshot(
	ctx context.Context,
	userID int,
	order *model.OrderDTO,
) (float64, *model.PromotionSnapshot) {
	if order == nil || order.PromotionCode == nil || order.PromotionCodeID == nil {
		return 0, nil
	}

	result, _ := r.applyPromotion(
		ctx,
		userID,
		order,
	)

	remakeCount := 0
	if order.RemakeCount != nil {
		remakeCount = *order.RemakeCount
	} else if order.LatestOrderItem != nil {
		remakeCount = order.LatestOrderItem.RemakeCount
	}

	isRemake := remakeCount > 0
	if order.RemakeType != nil && *order.RemakeType != "" {
		isRemake = true
	}

	promoSnapshot := &model.PromotionSnapshot{
		PromoCode:         result.Promotion.Code,
		DiscountType:      string(result.Promotion.DiscountType),
		DiscountValue:     result.Promotion.DiscountValue,
		DiscountAmount:    result.DiscountAmount,
		IsRemake:          isRemake,
		RemakeCount:       remakeCount,
		AppliedConditions: result.AppliedConditions,
		AppliedAt:         time.Now(),
	}

	return result.DiscountAmount, promoSnapshot
}

func (r *orderRepository) applyPromotionForSnapshot(
	ctx context.Context,
	userID int,
	order *model.OrderDTO,
	promoCodeString string,
) (*promotionApplyResult, error) {
	if strings.TrimSpace(promoCodeString) == "" {
		return nil, fmt.Errorf("promo_code_required")
	}
	if order == nil {
		return nil, fmt.Errorf("order_required")
	}

	promo, err := r.promotionRepo.GetByCode(ctx, promoCodeString)
	if err != nil {
		return nil, err
	}

	if !promo.IsActive {
		return nil, fmt.Errorf("promotion_inactive")
	}

	now := time.Now()
	if promo.StartAt.After(now) {
		return nil, fmt.Errorf("promotion_not_started")
	}
	if promo.EndAt.Before(now) {
		return nil, fmt.Errorf("promotion_expired")
	}

	if promo.TotalUsageLimit != nil && *promo.TotalUsageLimit != 0 {
		totalUsage, err := r.promotionRepo.CountTotalUsage(ctx, promo.ID)
		if err != nil {
			return nil, err
		}
		if totalUsage >= *promo.TotalUsageLimit {
			return nil, fmt.Errorf("promotion_total_usage_limit_reached")
		}
	}

	// if promo.UsagePerUser != nil && *promo.UsagePerUser != 0 {
	// 	userUsage, err := r.promotionRepo.CountUsageByUser(ctx, promo.ID, userID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if userUsage >= *promo.UsagePerUser {
	// 		return nil, fmt.Errorf("promotion_user_usage_limit_reached")
	// 	}
	// }

	orderCtx := r.buildOrderPromotionContext(order)

	scopeMatched, err := r.matchPromotionScopes(ctx, promo, userID, orderCtx)
	if err != nil {
		return nil, err
	}
	if !scopeMatched {
		return nil, fmt.Errorf("promotion_scope_not_matched")
	}

	appliedConditions, err := matchPromotionConditions(promo, orderCtx, now)
	if err != nil {
		return nil, err
	}

	discountAmount, err := calculatePromotionDiscount(promo, orderCtx)
	if err != nil {
		return nil, err
	}

	return &promotionApplyResult{
		DiscountAmount:    discountAmount,
		AppliedConditions: appliedConditions,
		Promo:             promo,
	}, nil
}

type orderPromotionContext struct {
	TotalPrice     float64
	IsRemake       bool
	RemakeCount    int
	RemakeReason   string
	OriginalTime   time.Time
	ProductIDs     []int
	ShippingAmount float64
	SellerID       int
}

func (r *orderRepository) buildOrderPromotionContext(order *model.OrderDTO) orderPromotionContext {
	totalPrice := r.orderItemProductRepo.CalculateTotalPrice(order.LatestOrderItem.Products)

	remakeCount := 0
	if order.RemakeCount != nil {
		remakeCount = *order.RemakeCount
	} else if order.LatestOrderItem != nil {
		remakeCount = order.LatestOrderItem.RemakeCount
	}

	isRemake := remakeCount > 0
	if order.RemakeType != nil && *order.RemakeType != "" {
		isRemake = true
	}

	remakeReason := ""
	if reason := utils.SafeGetStringPtr(order.CustomFields, "remake_reason"); reason != nil {
		remakeReason = *reason
	} else if order.LatestOrderItem != nil {
		if reason := utils.SafeGetStringPtr(order.LatestOrderItem.CustomFields, "remake_reason"); reason != nil {
			remakeReason = *reason
		}
	}

	originalTime := order.CreatedAt
	if originalTime.IsZero() && order.LatestOrderItem != nil {
		originalTime = order.LatestOrderItem.CreatedAt
	}

	productIDs := collectOrderProductIDs(order)

	shippingAmount := utils.SafeParseFloat(utils.SafeGet(order.CustomFields, "shipping_fee"))
	if shippingAmount == 0 {
		shippingAmount = utils.SafeParseFloat(utils.SafeGet(order.CustomFields, "shipping_cost"))
	}

	sellerID := 0
	if order.ClinicID != nil {
		sellerID = *order.ClinicID
	}

	return orderPromotionContext{
		TotalPrice:     *totalPrice,
		IsRemake:       isRemake,
		RemakeCount:    remakeCount,
		RemakeReason:   remakeReason,
		OriginalTime:   originalTime,
		ProductIDs:     productIDs,
		ShippingAmount: shippingAmount,
		SellerID:       sellerID,
	}
}

func collectOrderProductIDs(order *model.OrderDTO) []int {
	seen := map[int]struct{}{}
	var out []int
	if order.ProductID > 0 {
		seen[order.ProductID] = struct{}{}
		out = append(out, order.ProductID)
	}
	if order.LatestOrderItem != nil {
		for _, p := range order.LatestOrderItem.Products {
			if p == nil || p.ProductID <= 0 {
				continue
			}
			if _, ok := seen[p.ProductID]; ok {
				continue
			}
			seen[p.ProductID] = struct{}{}
			out = append(out, p.ProductID)
		}
	}
	return out
}

func (r *orderRepository) matchPromotionScopes(
	ctx context.Context,
	promo *generated.PromotionCode,
	userID int,
	orderCtx orderPromotionContext,
) (bool, error) {
	scopes := promo.Edges.Scopes
	if len(scopes) == 0 {
		return false, nil
	}

	hasCategoryScope := false
	for _, scope := range scopes {
		if scope.ScopeType == promotionmodel.PromotionScopeCategory {
			hasCategoryScope = true
			break
		}
	}

	var categoryIDs map[int]struct{}
	if hasCategoryScope && len(orderCtx.ProductIDs) > 0 {
		ids, err := r.loadPromotionCategoryIDs(ctx, orderCtx.ProductIDs)
		if err != nil {
			return false, err
		}
		categoryIDs = ids
	}

	for _, scope := range scopes {
		switch scope.ScopeType {
		case promotionmodel.PromotionScopeAll:
			return true, nil
		case promotionmodel.PromotionScopeUser:
			ids, err := parsePromotionIntList(scope.ScopeValue)
			if err != nil {
				return false, err
			}
			if containsPromotionInt(ids, userID) {
				return true, nil
			}
		case promotionmodel.PromotionScopeSeller:
			ids, err := parsePromotionIntList(scope.ScopeValue)
			if err != nil {
				return false, err
			}
			if orderCtx.SellerID != 0 && containsPromotionInt(ids, orderCtx.SellerID) {
				return true, nil
			}
		case promotionmodel.PromotionScopeProduct:
			ids, err := parsePromotionIntList(scope.ScopeValue)
			if err != nil {
				return false, err
			}
			if anyPromotionInSet(orderCtx.ProductIDs, ids) {
				return true, nil
			}
		case promotionmodel.PromotionScopeCategory:
			ids, err := parsePromotionIntList(scope.ScopeValue)
			if err != nil {
				return false, err
			}
			if anyPromotionInMap(ids, categoryIDs) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (r *orderRepository) loadPromotionCategoryIDs(ctx context.Context, productIDs []int) (map[int]struct{}, error) {
	if len(productIDs) == 0 {
		return map[int]struct{}{}, nil
	}

	products, err := r.db.Product.Query().
		Where(product.IDIn(productIDs...)).
		Select(product.FieldCategoryID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := map[int]struct{}{}
	for _, p := range products {
		if p == nil || p.CategoryID == nil {
			continue
		}
		out[*p.CategoryID] = struct{}{}
	}

	return out, nil
}

func matchPromotionConditions(
	promo *generated.PromotionCode,
	orderCtx orderPromotionContext,
	now time.Time,
) ([]string, error) {
	var applied []string
	for _, cond := range promo.Edges.Conditions {
		switch cond.ConditionType {
		case promotionmodel.PromotionConditionOrderIsRemake:
			if !orderCtx.IsRemake {
				return nil, fmt.Errorf("condition_order_is_remake_not_met")
			}
			applied = append(applied, string(cond.ConditionType))
		case promotionmodel.PromotionConditionRemakeCountLTE:
			value, err := parsePromotionIntValue(cond.ConditionValue)
			if err != nil {
				return nil, err
			}
			if orderCtx.RemakeCount > value {
				return nil, fmt.Errorf("condition_remake_count_lte_not_met")
			}
			applied = append(applied, string(cond.ConditionType))
		case promotionmodel.PromotionConditionRemakeWithinDays:
			value, err := parsePromotionIntValue(cond.ConditionValue)
			if err != nil {
				return nil, err
			}
			if orderCtx.OriginalTime.IsZero() {
				return nil, fmt.Errorf("condition_remake_within_days_not_met")
			}
			days := int(now.Sub(orderCtx.OriginalTime).Hours() / 24)
			if days > value {
				return nil, fmt.Errorf("condition_remake_within_days_not_met")
			}
			applied = append(applied, string(cond.ConditionType))
		case promotionmodel.PromotionConditionRemakeReason:
			values, err := parsePromotionStringList(cond.ConditionValue)
			if err != nil {
				return nil, err
			}
			if orderCtx.RemakeReason == "" || !containsPromotionString(values, orderCtx.RemakeReason) {
				return nil, fmt.Errorf("condition_remake_reason_not_met")
			}
			applied = append(applied, string(cond.ConditionType))
		default:
			return nil, fmt.Errorf("unsupported condition type: %s", cond.ConditionType)
		}
	}
	return applied, nil
}

func calculatePromotionDiscount(
	promo *generated.PromotionCode,
	orderCtx orderPromotionContext,
) (float64, error) {
	if promo.MinOrderValue != nil && *promo.MinOrderValue != 0 && orderCtx.TotalPrice < float64(*promo.MinOrderValue) {
		return 0, fmt.Errorf("min_order_value_not_met")
	}

	var discount float64
	switch promo.DiscountType {
	case promotionmodel.PromotionDiscountFixed:
		discount = float64(promo.DiscountValue)
	case promotionmodel.PromotionDiscountPercent:
		discount = orderCtx.TotalPrice * float64(promo.DiscountValue) / 100
	case promotionmodel.PromotionDiscountFreeShipping:
		discount = orderCtx.ShippingAmount
	default:
		return 0, fmt.Errorf("unsupported discount type: %s", promo.DiscountType)
	}

	if promo.MaxDiscountAmount != nil && *promo.MaxDiscountAmount != 0 && discount > float64(*promo.MaxDiscountAmount) {
		discount = float64(*promo.MaxDiscountAmount)
	}
	if discount < 0 {
		discount = 0
	}
	if discount > orderCtx.TotalPrice {
		discount = orderCtx.TotalPrice
	}

	return discount, nil
}

func parsePromotionIntValue(raw json.RawMessage) (int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, errors.New("missing value")
	}
	var val int
	if err := json.Unmarshal(raw, &val); err == nil {
		return val, nil
	}
	var fval float64
	if err := json.Unmarshal(raw, &fval); err == nil {
		return int(fval), nil
	}
	var sval string
	if err := json.Unmarshal(raw, &sval); err == nil {
		i, convErr := strconv.Atoi(sval)
		if convErr == nil {
			return i, nil
		}
	}
	return 0, errors.New("invalid int value")
}

func parsePromotionIntList(raw json.RawMessage) ([]int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var single int
	if err := json.Unmarshal(raw, &single); err == nil {
		return []int{single}, nil
	}
	var list []int
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, nil
	}
	var anyList []any
	if err := json.Unmarshal(raw, &anyList); err == nil {
		out := make([]int, 0, len(anyList))
		for _, item := range anyList {
			switch v := item.(type) {
			case float64:
				out = append(out, int(v))
			case int:
				out = append(out, v)
			case string:
				i, err := strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
				out = append(out, i)
			default:
				return nil, errors.New("invalid int list item")
			}
		}
		return out, nil
	}
	return nil, errors.New("invalid int list")
}

func parsePromotionStringList(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var single string
	if err := json.Unmarshal(raw, &single); err == nil {
		return []string{single}, nil
	}
	var list []string
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, nil
	}
	var anyList []any
	if err := json.Unmarshal(raw, &anyList); err == nil {
		out := make([]string, 0, len(anyList))
		for _, item := range anyList {
			switch v := item.(type) {
			case string:
				out = append(out, v)
			default:
				return nil, errors.New("invalid string list item")
			}
		}
		return out, nil
	}
	return nil, errors.New("invalid string list")
}

func containsPromotionInt(list []int, target int) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func containsPromotionString(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func anyPromotionInSet(orderIDs []int, allowed []int) bool {
	for _, id := range orderIDs {
		if containsPromotionInt(allowed, id) {
			return true
		}
	}
	return false
}

func anyPromotionInMap(ids []int, allowed map[int]struct{}) bool {
	if len(allowed) == 0 {
		return false
	}
	for _, id := range ids {
		if _, ok := allowed[id]; ok {
			return true
		}
	}
	return false
}
