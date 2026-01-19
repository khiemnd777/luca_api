package service

import (
	"context"
	"errors"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

func (s *promotionService) CreatePromotion(
	ctx context.Context,
	input *model.CreatePromotionInput,
) (*generated.PromotionCode, error) {
	if input == nil {
		return nil, errors.New("input is required")
	}
	return s.repo.CreatePromotion(ctx, input)
}

func (s *promotionService) UpdatePromotion(
	ctx context.Context,
	id int,
	input *model.UpdatePromotionInput,
) (*generated.PromotionCode, error) {
	if input == nil {
		return nil, errors.New("input is required")
	}
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.repo.UpdatePromotion(ctx, id, input)
}

func (s *promotionService) DeletePromotion(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return s.repo.DeletePromotion(ctx, id)
}

func (s *promotionService) GetPromotionByID(
	ctx context.Context,
	id int,
) (*generated.PromotionCode, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.repo.GetPromotionByID(ctx, id)
}

func (s *promotionService) ListPromotions(
	ctx context.Context,
	query table.TableQuery,
) (table.TableListResult[model.PromotionCodeDTO], error) {
	return s.repo.ListPromotions(ctx, query)
}
