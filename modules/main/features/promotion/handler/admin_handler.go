package handler

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	promotionservice "github.com/khiemnd777/andy_api/modules/main/features/promotion/service"
	"github.com/khiemnd777/andy_api/shared/app"
	"github.com/khiemnd777/andy_api/shared/app/client_error"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotioncode"
	"github.com/khiemnd777/andy_api/shared/middleware/rbac"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type PromotionAdminHandler struct {
	svc  promotionservice.PromotionService
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewPromotionAdminHandler(
	svc promotionservice.PromotionService,
	deps *module.ModuleDeps[config.ModuleConfig],
) *PromotionAdminHandler {
	return &PromotionAdminHandler{svc: svc, deps: deps}
}

func (h *PromotionAdminHandler) RegisterRoutes(router fiber.Router) {
	app.RouterGet(router, "/:dept_id<int>/promotion/list", h.List)
	app.RouterGet(router, "/:dept_id<int>/promotion/:id<int>", h.GetByID)
	app.RouterPost(router, "/:dept_id<int>/promotion", h.Create)
	app.RouterPut(router, "/:dept_id<int>/promotion/:id<int>", h.Update)
	app.RouterDelete(router, "/:dept_id<int>/promotion/:id<int>", h.Delete)
}

type promotionCreatePayload struct {
	Code              string                     `json:"code"`
	DiscountType      promotioncode.DiscountType `json:"discount_type"`
	DiscountValue     int                        `json:"discount_value"`
	MaxDiscountAmount *int                       `json:"max_discount_amount"`
	MinOrderValue     *int                       `json:"min_order_value"`
	TotalUsageLimit   *int                       `json:"total_usage_limit"`
	UsagePerUser      *int                       `json:"usage_per_user"`
	StartAt           *time.Time                 `json:"start_at"`
	EndAt             *time.Time                 `json:"end_at"`
	IsActive          bool                       `json:"is_active"`
}

type promotionUpdatePayload struct {
	DiscountType      promotioncode.DiscountType `json:"discount_type"`
	DiscountValue     int                        `json:"discount_value"`
	MaxDiscountAmount *int                       `json:"max_discount_amount"`
	MinOrderValue     *int                       `json:"min_order_value"`
	TotalUsageLimit   *int                       `json:"total_usage_limit"`
	UsagePerUser      *int                       `json:"usage_per_user"`
	StartAt           *time.Time                 `json:"start_at"`
	EndAt             *time.Time                 `json:"end_at"`
	IsActive          *bool                      `json:"is_active"`
}

func (h *PromotionAdminHandler) List(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "promotion.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	q := table.ParseTableQuery(c, 20)
	items, err := h.svc.ListPromotions(c.UserContext(), q)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(items)
}

func (h *PromotionAdminHandler) GetByID(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "promotion.view"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}

	item, err := h.svc.GetPromotionByID(c.UserContext(), id)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(item)
}

func (h *PromotionAdminHandler) Create(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "promotion.create"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	var payload promotionCreatePayload
	if err := c.BodyParser(&payload); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}
	if strings.TrimSpace(payload.Code) == "" {
		return client_error.ResponseError(c, fiber.StatusBadRequest, nil, "code is required")
	}

	input := &model.CreatePromotionInput{
		Code:              payload.Code,
		DiscountType:      payload.DiscountType,
		DiscountValue:     payload.DiscountValue,
		MaxDiscountAmount: payload.MaxDiscountAmount,
		MinOrderValue:     payload.MinOrderValue,
		TotalUsageLimit:   payload.TotalUsageLimit,
		UsagePerUser:      payload.UsagePerUser,
		StartAt:           payload.StartAt,
		EndAt:             payload.EndAt,
		IsActive:          payload.IsActive,
	}

	item, err := h.svc.CreatePromotion(c.UserContext(), input)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *PromotionAdminHandler) Update(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "promotion.update"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}

	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}

	var payload promotionUpdatePayload
	if err := c.BodyParser(&payload); err != nil {
		return client_error.ResponseError(c, fiber.StatusBadRequest, err, "invalid body")
	}

	input := &model.UpdatePromotionInput{
		DiscountType:      payload.DiscountType,
		DiscountValue:     payload.DiscountValue,
		MaxDiscountAmount: payload.MaxDiscountAmount,
		MinOrderValue:     payload.MinOrderValue,
		TotalUsageLimit:   payload.TotalUsageLimit,
		UsagePerUser:      payload.UsagePerUser,
		StartAt:           payload.StartAt,
		EndAt:             payload.EndAt,
		IsActive:          payload.IsActive,
	}

	item, err := h.svc.UpdatePromotion(c.UserContext(), id, input)
	if err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(item)
}

func (h *PromotionAdminHandler) Delete(c *fiber.Ctx) error {
	if err := rbac.GuardAnyPermission(c, h.deps.Ent.(*generated.Client), "promotion.delete"); err != nil {
		return client_error.ResponseError(c, fiber.StatusForbidden, err, err.Error())
	}
	id, _ := utils.GetParamAsInt(c, "id")
	if id <= 0 {
		return client_error.ResponseError(c, fiber.StatusNotFound, nil, "invalid id")
	}

	if err := h.svc.DeletePromotion(c.UserContext(), id); err != nil {
		return client_error.ResponseError(c, fiber.StatusInternalServerError, err, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
