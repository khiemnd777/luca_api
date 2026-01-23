package engine

import (
	"fmt"

	promotionmodel "github.com/khiemnd777/andy_api/modules/main/features/promotion/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/logger"
)

func (e *Engine) calculateDiscount(
	promo *generated.PromotionCode,
	orderCtx OrderContext,
) (float64, error) {

	// ===== Min order value check
	if promo.MinOrderValue != nil &&
		*promo.MinOrderValue != 0 &&
		orderCtx.TotalPrice < float64(*promo.MinOrderValue) {

		logger.Info(
			"promotion discount rejected: min order value not met",
			"promo_code", promo.Code,
			"total_price", orderCtx.TotalPrice,
			"min_order_value", *promo.MinOrderValue,
		)

		return 0, PromotionApplyError{Reason: ReasonMinOrderValueNotMet}
	}

	var discount float64

	// ===== Discount calculation
	switch promo.DiscountType {

	case promotionmodel.PromotionDiscountFixed:
		discount = float64(promo.DiscountValue)

		logger.Debug(
			"promotion fixed discount calculated",
			"promo_code", promo.Code,
			"discount_value", promo.DiscountValue,
		)

	case promotionmodel.PromotionDiscountPercent:
		discount = orderCtx.TotalPrice * float64(promo.DiscountValue) / 100

		logger.Debug(
			"promotion percent discount calculated",
			"promo_code", promo.Code,
			"percent", promo.DiscountValue,
			"base_price", orderCtx.TotalPrice,
			"raw_discount", discount,
		)

	case promotionmodel.PromotionDiscountFreeShipping:
		discount = orderCtx.ShippingAmount

		logger.Debug(
			"promotion free shipping discount applied",
			"promo_code", promo.Code,
			"shipping_amount", orderCtx.ShippingAmount,
		)

	default:
		logger.Error(
			"unsupported promotion discount type",
			"promo_code", promo.Code,
			"discount_type", promo.DiscountType,
		)
		return 0, fmt.Errorf("unsupported discount type: %s", promo.DiscountType)
	}

	// ===== Max discount cap
	if promo.MaxDiscountAmount != nil &&
		*promo.MaxDiscountAmount != 0 &&
		discount > float64(*promo.MaxDiscountAmount) {

		logger.Debug(
			"promotion discount capped by max discount amount",
			"promo_code", promo.Code,
			"raw_discount", discount,
			"max_discount_amount", *promo.MaxDiscountAmount,
		)

		discount = float64(*promo.MaxDiscountAmount)
	}

	// ===== Normalize discount
	if discount < 0 {
		logger.Warn(
			"promotion discount normalized: negative discount",
			"promo_code", promo.Code,
			"discount", discount,
		)
		discount = 0
	}

	if discount > orderCtx.TotalPrice {
		logger.Debug(
			"promotion discount normalized: capped by total price",
			"promo_code", promo.Code,
			"discount", discount,
			"total_price", orderCtx.TotalPrice,
		)
		discount = orderCtx.TotalPrice
	}

	logger.Info(
		"promotion discount calculated",
		"promo_code", promo.Code,
		"discount_type", promo.DiscountType,
		"discount_amount", discount,
	)

	return discount, nil
}
