package contextbuilder

import (
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/promotion/engine"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type OrderItemPriceCalculator interface {
	CalculateTotalPrice(
		products []*model.OrderItemProductDTO,
	) *float64
}

type Builder struct {
	priceCalc OrderItemPriceCalculator
}

func NewBuilder(
	priceCalc OrderItemPriceCalculator,
) *Builder {
	return &Builder{
		priceCalc: priceCalc,
	}
}

func (b *Builder) BuildFromOrderDTO(
	order *model.OrderDTO,
) engine.OrderContext {

	totalPrice := b.priceCalc.
		CalculateTotalPrice(order.LatestOrderItem.Products)

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
		if reason := utils.SafeGetStringPtr(
			order.LatestOrderItem.CustomFields,
			"remake_reason",
		); reason != nil {
			remakeReason = *reason
		}
	}

	originalTime := order.CreatedAt
	if originalTime.IsZero() && order.LatestOrderItem != nil {
		originalTime = order.LatestOrderItem.CreatedAt
	}

	productIDs := collectOrderProductIDs(order)

	shippingAmount := utils.SafeParseFloat(
		utils.SafeGet(order.CustomFields, "shipping_fee"),
	)
	if shippingAmount == 0 {
		shippingAmount = utils.SafeParseFloat(
			utils.SafeGet(order.CustomFields, "shipping_cost"),
		)
	}

	sellerID := 0
	if order.ClinicID != nil {
		sellerID = *order.ClinicID
	}

	clinicID := 0
	if order.ClinicID != nil {
		clinicID = *order.ClinicID
	}

	refUserID := 0
	if order.RefUserID != nil {
		refUserID = *order.RefUserID
	}

	return engine.OrderContext{
		TotalPrice:     *totalPrice,
		IsRemake:       isRemake,
		RemakeCount:    remakeCount,
		RemakeReason:   remakeReason,
		OriginalTime:   originalTime,
		ProductIDs:     productIDs,
		ShippingAmount: shippingAmount,
		SellerID:       sellerID,
		ClinicID:       clinicID,
		RefUserID:      refUserID,
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
