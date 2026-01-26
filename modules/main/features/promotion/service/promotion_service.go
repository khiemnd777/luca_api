package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	orderrepo "github.com/khiemnd777/andy_api/modules/main/features/order/repository"
	"github.com/khiemnd777/andy_api/modules/main/features/promotion/contextbuilder"
	"github.com/khiemnd777/andy_api/modules/main/features/promotion/engine"
	promotionmodel "github.com/khiemnd777/andy_api/modules/main/features/promotion/model"
	"github.com/khiemnd777/andy_api/modules/main/features/promotion/repository"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/product"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type PromotionApplyResult struct {
	DiscountAmount    float64                  `json:"discount_amount"`
	FinalPrice        float64                  `json:"final_price"`
	AppliedConditions []string                 `json:"applied_conditions"`
	PromoCode         string                   `json:"promo_code"`
	Promotion         *generated.PromotionCode `json:"-"`
}

type PromotionApplyError struct {
	Reason string
}

func (e PromotionApplyError) Error() string {
	return e.Reason
}

func IsPromotionApplyError(err error) (string, bool) {
	var perr PromotionApplyError
	if errors.As(err, &perr) {
		return perr.Reason, true
	}
	return "", false
}

type PromotionService interface {
	ApplyPromotion(ctx context.Context, userID *int, order *model.OrderDTO, promoCodeString string) (*engine.PromotionApplyResult, error)
	ApplyPromotionAndSnapshot(ctx context.Context, userID *int, order *model.OrderDTO, promoCodeString string) (*engine.PromotionApplyResult, *model.PromotionSnapshot, error)
	GetPromotionCodesInUsageByOrderID(ctx context.Context, orderID int64) ([]model.PromotionCodeDTO, error)
}

type promotionService struct {
	repo                 repository.PromotionRepository
	orderItemProductRepo orderrepo.OrderItemProductRepository
	promoengine          *engine.Engine
	promoctxbuilder      *contextbuilder.Builder
	promoguard           engine.PromotionGuard
	deps                 *module.ModuleDeps[config.ModuleConfig]
}

func NewPromotionService(repo repository.PromotionRepository, deps *module.ModuleDeps[config.ModuleConfig]) PromotionService {
	orderItemProductRepo := orderrepo.NewOrderItemProductRepository(deps.Ent.(*generated.Client))
	promoengine := engine.NewEngine(deps)
	promoctxbuilder := contextbuilder.NewBuilder(orderItemProductRepo)
	promoguard := engine.NewGuard(repo)
	return &promotionService{
		repo:                 repo,
		orderItemProductRepo: orderItemProductRepo,
		promoengine:          promoengine,
		promoctxbuilder:      promoctxbuilder,
		promoguard:           promoguard,
		deps:                 deps,
	}
}

func (s *promotionService) ApplyPromotionV1(
	ctx context.Context,
	userID int,
	order *model.OrderDTO,
	promoCodeString string,
) (*PromotionApplyResult, error) {

	if strings.TrimSpace(promoCodeString) == "" {
		return nil, PromotionApplyError{Reason: engine.ReasonPromoCodeRequired}
	}
	if order == nil {
		return nil, PromotionApplyError{Reason: engine.ReasonOrderRequired}
	}

	promo, err := s.repo.GetByCode(ctx, promoCodeString)

	if err != nil {
		if generated.IsNotFound(err) {
			return nil, PromotionApplyError{Reason: engine.ReasonPromotionNotFound}
		}
		return nil, err
	}

	if !promo.IsActive {
		return nil, PromotionApplyError{Reason: engine.ReasonPromotionInactive}
	}

	now := time.Now()
	if promo.StartAt.After(now) {
		return nil, PromotionApplyError{Reason: engine.ReasonPromotionNotStarted}
	}
	if promo.EndAt != nil && promo.EndAt.Before(now) {
		return nil, PromotionApplyError{Reason: engine.ReasonPromotionExpired}
	}

	if promo.TotalUsageLimit != nil && *promo.TotalUsageLimit != 0 {
		totalUsage, err := s.repo.CountTotalUsage(ctx, promo.ID)
		if err != nil {
			return nil, err
		}
		if totalUsage >= *promo.TotalUsageLimit {
			return nil, PromotionApplyError{Reason: engine.ReasonPromotionTotalUsageLimitReached}
		}
	}

	// if promo.UsagePerUser != nil && *promo.UsagePerUser != 0 {
	// 	userUsage, err := s.repo.CountUsageByUser(ctx, promo.ID, userID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if userUsage >= *promo.UsagePerUser {
	// 		return nil, PromotionApplyError{Reason: engine.ReasonPromotionUserUsageLimitReached}
	// 	}
	// }

	orderCtx := s.buildOrderPromotionContext(order)

	scopeMatched, err := s.matchScopes(ctx, promo, userID, orderCtx)
	if err != nil {
		return nil, err
	}
	if !scopeMatched {
		return nil, PromotionApplyError{Reason: engine.ReasonPromotionScopeNotMatched}
	}

	appliedConditions, err := s.matchConditions(promo, orderCtx, now)
	if err != nil {
		return nil, err
	}

	discountAmount, err := s.calculateDiscount(promo, orderCtx)
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

func (s *promotionService) ApplyPromotion(
	ctx context.Context,
	userID *int,
	order *model.OrderDTO,
	promoCodeString string,
) (*engine.PromotionApplyResult, error) {

	orderCtx := s.promoctxbuilder.BuildFromOrderDTO(order)
	promo, err := s.repo.GetByCode(ctx, promoCodeString)
	result, err := s.promoengine.Apply(
		ctx,
		promo,
		s.promoguard,
		order.RefUserID,
		order.ID,
		orderCtx,
		time.Now(),
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *promotionService) ApplyPromotionAndSnapshotV1(
	ctx context.Context,
	userID int,
	order *model.OrderDTO,
	promoCodeString string,
) (*PromotionApplyResult, *model.PromotionSnapshot, error) {
	result, err := s.ApplyPromotionV1(ctx, userID, order, promoCodeString)
	if err != nil {
		return nil, nil, err
	}
	if order == nil || order.ID == 0 {
		return nil, nil, errors.New("order is required")
	}

	orderCtx := s.buildOrderPromotionContext(order)

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

	if err := s.repo.CreatePromotionUsageFromSnapshot(
		ctx,
		nil,
		result.Promotion.ID,
		order.ID,
		order.RefUserID,
		snapshot,
	); err != nil {
		return nil, nil, err
	}

	return result, snapshot, nil
}

func (s *promotionService) ApplyPromotionAndSnapshot(
	ctx context.Context,
	userID *int,
	order *model.OrderDTO,
	promoCodeString string,
) (*engine.PromotionApplyResult, *model.PromotionSnapshot, error) {
	result, err := s.ApplyPromotion(ctx, order.RefUserID, order, promoCodeString)
	if err != nil {
		return nil, nil, err
	}
	if order == nil || order.ID == 0 {
		return nil, nil, errors.New("order is required")
	}

	orderCtx := s.buildOrderPromotionContext(order)

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

	if err := s.repo.UpsertPromotionUsageFromSnapshot(ctx, nil, result.Promotion.ID, order.ID, order.RefUserID, snapshot); err != nil {
		return nil, nil, err
	}

	return result, snapshot, nil
}

func (s *promotionService) GetPromotionCodesInUsageByOrderID(ctx context.Context, orderID int64) ([]model.PromotionCodeDTO, error) {
	return s.repo.GetPromotionCodesInUsageByOrderID(ctx, orderID)
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

func (s *promotionService) buildOrderPromotionContext(order *model.OrderDTO) orderPromotionContext {
	totalPrice := s.orderItemProductRepo.CalculateTotalPrice(order.LatestOrderItem.Products)

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

func (s *promotionService) matchScopes(
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
		ids, err := s.loadCategoryIDs(ctx, orderCtx.ProductIDs)
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
			ids, err := parseIntList(scope.ScopeValue)
			if err != nil {
				return false, err
			}
			if containsInt(ids, userID) {
				return true, nil
			}
		case promotionmodel.PromotionScopeSeller:
			ids, err := parseIntList(scope.ScopeValue)
			if err != nil {
				return false, err
			}
			if orderCtx.SellerID != 0 && containsInt(ids, orderCtx.SellerID) {
				return true, nil
			}
		case promotionmodel.PromotionScopeProduct:
			ids, err := parseIntList(scope.ScopeValue)
			if err != nil {
				return false, err
			}
			if anyInSet(orderCtx.ProductIDs, ids) {
				return true, nil
			}
		case promotionmodel.PromotionScopeCategory:
			ids, err := parseIntList(scope.ScopeValue)
			if err != nil {
				return false, err
			}
			if anyInMap(ids, categoryIDs) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (s *promotionService) loadCategoryIDs(ctx context.Context, productIDs []int) (map[int]struct{}, error) {
	logger.Debug("loadCategoryIDs: start",
		"productIDs", productIDs,
		"productCount", len(productIDs),
	)

	client, ok := s.deps.Ent.(*generated.Client)
	if !ok || client == nil {
		logger.Debug("loadCategoryIDs: invalid ent client")
		return nil, errors.New("invalid ent client")
	}

	if len(productIDs) == 0 {
		logger.Debug("loadCategoryIDs: empty productIDs, return empty result")
		return map[int]struct{}{}, nil
	}

	products, err := client.Product.Query().
		Where(product.IDIn(productIDs...)).
		Select(product.FieldCategoryID).
		All(ctx)
	if err != nil {
		logger.Debug("loadCategoryIDs: query failed",
			"error", err,
			"productIDs", productIDs,
		)
		return nil, err
	}

	logger.Debug("loadCategoryIDs: products loaded",
		"productsCount", len(products),
	)

	out := map[int]struct{}{}
	for _, p := range products {
		if p == nil {
			logger.Debug("loadCategoryIDs: skip nil product")
			continue
		}
		if p.CategoryID == nil {
			logger.Debug("loadCategoryIDs: product has nil CategoryID",
				"product", p,
			)
			continue
		}
		out[*p.CategoryID] = struct{}{}
	}

	logger.Debug("loadCategoryIDs: result",
		"uniqueCategoryCount", len(out),
		"categoryIDs", func() []int {
			ids := make([]int, 0, len(out))
			for id := range out {
				ids = append(ids, id)
			}
			return ids
		}(),
	)

	return out, nil
}

func (s *promotionService) matchConditions(
	promo *generated.PromotionCode,
	orderCtx orderPromotionContext,
	now time.Time,
) ([]string, error) {
	var applied []string
	for _, cond := range promo.Edges.Conditions {
		switch cond.ConditionType {
		case promotionmodel.PromotionConditionOrderIsRemake:
			if !orderCtx.IsRemake {
				return nil, PromotionApplyError{Reason: engine.ReasonConditionOrderIsRemakeNotMet}
			}
			applied = append(applied, string(cond.ConditionType))
		case promotionmodel.PromotionConditionRemakeCountLTE:
			value, err := parseIntValue(cond.ConditionValue)
			if err != nil {
				return nil, err
			}
			if orderCtx.RemakeCount > value {
				return nil, PromotionApplyError{Reason: engine.ReasonConditionRemakeCountLTENotMet}
			}
			applied = append(applied, string(cond.ConditionType))
		case promotionmodel.PromotionConditionRemakeWithinDays:
			value, err := parseIntValue(cond.ConditionValue)
			if err != nil {
				return nil, err
			}
			if orderCtx.OriginalTime.IsZero() {
				return nil, PromotionApplyError{Reason: engine.ReasonConditionRemakeWithinDaysNotMet}
			}
			days := int(now.Sub(orderCtx.OriginalTime).Hours() / 24)
			if days > value {
				return nil, PromotionApplyError{Reason: engine.ReasonConditionRemakeWithinDaysNotMet}
			}
			applied = append(applied, string(cond.ConditionType))
		case promotionmodel.PromotionConditionRemakeReason:
			values, err := parseStringList(cond.ConditionValue)
			if err != nil {
				return nil, err
			}
			if orderCtx.RemakeReason == "" || !containsString(values, orderCtx.RemakeReason) {
				return nil, PromotionApplyError{Reason: engine.ReasonConditionRemakeReasonNotMet}
			}
			applied = append(applied, string(cond.ConditionType))
		default:
			return nil, fmt.Errorf("unsupported condition type: %s", cond.ConditionType)
		}
	}
	return applied, nil
}

func (s *promotionService) calculateDiscount(
	promo *generated.PromotionCode,
	orderCtx orderPromotionContext,
) (float64, error) {
	if promo.MinOrderValue != nil && *promo.MinOrderValue != 0 && orderCtx.TotalPrice < float64(*promo.MinOrderValue) {
		return 0, PromotionApplyError{Reason: engine.ReasonMinOrderValueNotMet}
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

func parseIntValue(raw json.RawMessage) (int, error) {
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

func parseIntList(raw json.RawMessage) ([]int, error) {
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

func parseStringList(raw json.RawMessage) ([]string, error) {
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

func containsInt(list []int, target int) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func containsString(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func anyInSet(orderIDs []int, allowed []int) bool {
	for _, id := range orderIDs {
		if containsInt(allowed, id) {
			return true
		}
	}
	return false
}

func anyInMap(ids []int, allowed map[int]struct{}) bool {
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
