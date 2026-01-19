package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
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
}

func (h *PromotionHandler) Validate(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "promotion.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	payload, err := app.ParseBody[struct {
		PromoCode string `json:"promo_code"`
		OrderID   int64  `json:"order_id"`
	}](c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	if payload.OrderID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order id")
	}

	userID, ok := utils.GetUserIDInt(c)
	if !ok {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, nil, "unauthorized")
	}

	order, err := h.orderSvc.GetByID(c.UserContext(), payload.OrderID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	result, err := h.svc.ApplyPromotion(c.UserContext(), userID, order, payload.PromoCode)
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
		PromoCode string `json:"promo_code"`
		OrderID   int64  `json:"order_id"`
	}](c)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	if payload.OrderID <= 0 {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "invalid order id")
	}

	userID, ok := utils.GetUserIDInt(c)
	if !ok {
		return client_error.ResponseError(c, fiber.StatusUnauthorized, nil, "unauthorized")
	}

	order, err := h.orderSvc.GetByID(c.UserContext(), payload.OrderID)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}

	result, snapshot, err := h.svc.ApplyPromotionAndSnapshot(c.UserContext(), userID, order, payload.PromoCode)
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
