package engine

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/features/promotion/repository"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
)

type PromotionGuard interface {
	// Validate promo existence & lifecycle
	EnsureValidPromo(
		ctx context.Context,
		promo *generated.PromotionCode,
		now time.Time,
		orderID int64,
	) error

	// Check usage limits
	CheckUsage(
		ctx context.Context,
		promo *generated.PromotionCode,
		userID int,
		orderID int64,
	) error
}

type Guard struct {
	repo repository.PromotionRepository
}

func NewGuard(repo repository.PromotionRepository) PromotionGuard {
	return &Guard{repo: repo}
}

func (g *Guard) EnsureValidPromo(
	ctx context.Context,
	promo *generated.PromotionCode,
	now time.Time,
	orderID int64,
) error {

	used, err := g.repo.ExistsUsageByOrderID(ctx, promo.ID, orderID)
	if err != nil {
		return err
	}
	if used {
		return nil
	}

	if promo == nil {
		return PromotionApplyError{Reason: "promotion_not_found"}
	}

	if !promo.IsActive {
		return PromotionApplyError{Reason: "promotion_inactive"}
	}

	if promo.StartAt.After(now) {
		return PromotionApplyError{Reason: "promotion_not_started"}
	}

	if promo.EndAt.Before(now) {
		return PromotionApplyError{Reason: "promotion_expired"}
	}

	return nil
}

func (g *Guard) CheckUsage(
	ctx context.Context,
	promo *generated.PromotionCode,
	userID int,
	orderID int64,
) error {

	used, err := g.repo.ExistsUsageByOrderID(ctx, promo.ID, orderID)
	if err != nil {
		return err
	}
	if used {
		return nil
	}

	// ===== TOTAL USAGE LIMIT =====
	if promo.TotalUsageLimit != nil && *promo.TotalUsageLimit != 0 {
		total, err := g.repo.CountTotalUsage(ctx, promo.ID)
		if err != nil {
			return err
		}
		if total >= *promo.TotalUsageLimit {
			return PromotionApplyError{
				Reason: "promotion_total_usage_limit_reached",
			}
		}
	}

	// ===== USER USAGE LIMIT (optional – giữ y hệt code cũ) =====
	/*
		if promo.UsagePerUser != nil && *promo.UsagePerUser != 0 {
			userUsage, err := g.repo.CountUsageByUser(ctx, promo.ID, userID)
			if err != nil {
				return err
			}
			if userUsage >= *promo.UsagePerUser {
				return engine.PromotionApplyError{
					Reason: "promotion_user_usage_limit_reached",
				}
			}
		}
	*/

	return nil
}
