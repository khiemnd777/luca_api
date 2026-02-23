package engine

import (
	"context"
	"errors"
	"sort"

	promotionmodel "github.com/khiemnd777/andy_api/modules/main/features/promotion/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/product"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func (e *Engine) matchScopes(
	ctx context.Context,
	promo *generated.PromotionCode,
	userID *int,
	orderCtx OrderContext,
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
		ids, err := e.loadCategoryIDs(ctx, orderCtx.ProductIDs)
		if err != nil {
			return false, err
		}
		categoryIDs = ids
	}

	logger.Debug("matchScopes: loaded categoryIDs",
		"productIDs", orderCtx.ProductIDs,
		"categoryIDs", mapKeys(categoryIDs),
	)

	for _, scope := range scopes {
		switch scope.ScopeType {

		case promotionmodel.PromotionScopeAll:
			return true, nil

		case promotionmodel.PromotionScopeUser:
			ids, err := parseIntList(scope.ScopeValue)
			if err != nil {
				return false, err
			}
			if userID == nil {
				return false, nil
			}
			if containsInt(ids, *userID) {
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

		case promotionmodel.PromotionScopeClinic:
			ids, err := parseIntList(scope.ScopeValue)
			if err != nil {
				return false, err
			}
			if orderCtx.ClinicID != 0 && containsInt(ids, orderCtx.ClinicID) {
				return true, nil
			}

		case promotionmodel.PromotionScopeStaff:
			ids, err := parseIntList(scope.ScopeValue)
			if err != nil {
				return false, err
			}
			if orderCtx.RefUserID != 0 && containsInt(ids, orderCtx.RefUserID) {
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
			logger.Debug("matchScopes: checking category scope",
				"scopeValue", scope.ScopeValue,
				"parsedIDs", ids,
				"categoryIDs", mapKeys(categoryIDs),
			)
			if anyInMap(ids, categoryIDs) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (e *Engine) loadCategoryIDs(ctx context.Context, productIDs []int) (map[int]struct{}, error) {
	if len(productIDs) == 0 {
		return map[int]struct{}{}, nil
	}
	logger.Debug("loadCategoryIDs: start", "productIDs", productIDs)
	client, ok := e.deps.Ent.(*generated.Client)
	if !ok || client == nil {
		return nil, errors.New("invalid ent client")
	}

	categoryIDs, err := client.Product.Query().
		Where(
			product.IDIn(productIDs...),
			product.CategoryIDNotNil(),
		).
		Select(product.FieldCategoryID).
		Ints(ctx)
	if err != nil {
		return nil, err
	}

	out := make(map[int]struct{}, len(categoryIDs))
	for _, id := range categoryIDs {
		out[id] = struct{}{}
	}

	logger.Debug("loadCategoryIDs: done", "categoryIDs", out)
	return out, nil
}

func mapKeys(m map[int]struct{}) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
