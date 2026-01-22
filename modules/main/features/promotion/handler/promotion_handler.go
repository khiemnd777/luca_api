package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/order/service"
	promotionservice "github.com/khiemnd777/andy_api/modules/main/features/promotion/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type PromotionHandler struct {
	svc      promotionservice.PromotionService
	orderSvc service.OrderService
	deps     *module.ModuleDeps[config.ModuleConfig]
}

func NewPromotionHandler(
	svc promotionservice.PromotionService,
	orderSvc service.OrderService,
	deps *module.ModuleDeps[config.ModuleConfig],
) *PromotionHandler {
	return &PromotionHandler{svc: svc, orderSvc: orderSvc, deps: deps}
}

func (h *PromotionHandler) RegisterRoutes(router fiber.Router) {
	app.RouterPost(router, "/:dept_id<int>/promotions/validate", h.Validate)
	app.RouterPost(router, "/:dept_id<int>/promotions/apply", h.Apply)
	app.RouterGet(router, "/:dept_id<int>/order/:order_id<int>/promotions", h.GetPromotionCodesInUsageByOrderID)
}

func (h *PromotionHandler) Validate(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "promotion.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	payload, err := app.ParseBody[struct {
		PromoCode string          `json:"promo_code"`
		Order     *model.OrderDTO `json:"order"`
	}](c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	if payload.Order == nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order")
	}

	userID, ok := utils.GetUserIDInt(c)
	if !ok {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, nil, "unauthorized")
	}

	result, err := h.svc.ApplyPromotion(c.UserContext(), userID, payload.Order, payload.PromoCode)
	if err != nil {
		if reason, ok := promotionservice.IsPromotionApplyError(err); ok {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"valid":  false,
				"reason": reason,
			})
		}
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"valid":           true,
		"discount_amount": result.DiscountAmount,
		"final_price":     result.FinalPrice,
	})
}

func (h *PromotionHandler) Apply(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "promotion.update"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	payload, err := app.ParseBody[struct {
		PromoCode string          `json:"promo_code"`
		Order     *model.OrderDTO `json:"order"`
	}](c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	if payload.Order == nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order")
	}

	userID, ok := utils.GetUserIDInt(c)
	if !ok {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, nil, "unauthorized")
	}

	if payload.Order.ID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order id")
	}

	result, snapshot, err := h.svc.ApplyPromotionAndSnapshot(c.UserContext(), userID, payload.Order, payload.PromoCode)
	if err != nil {
		if reason, ok := promotionservice.IsPromotionApplyError(err); ok {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"success": false,
				"reason":  reason,
			})
		}
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":          true,
		"applied_discount": result.DiscountAmount,
		"promo_snapshot":   snapshot,
	})
}

func (h *PromotionHandler) GetPromotionCodesInUsageByOrderID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "promotion.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	orderID, _ := utils.GetParamAsInt(c, "order_id")
	if orderID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order id")
	}

	items, err := h.svc.GetPromotionCodesInUsageByOrderID(c.UserContext(), orderID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(items)
}
