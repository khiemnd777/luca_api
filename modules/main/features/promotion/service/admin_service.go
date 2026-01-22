package service

import (
	"context"
	"errors"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

func (s *promotionService) CreatePromotion(
	ctx context.Context,
	input *model.CreatePromotionInput,
) (*model.PromotionCodeDTO, error) {
	if input == nil {
		return nil, errors.New("input is required")
	}
	if err := validatePromotionInput(input); err != nil {
		return nil, err
	}
	item, err := s.repo.CreatePromotion(ctx, input)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *promotionService) UpdatePromotion(
	ctx context.Context,
	id int,
	input *model.UpdatePromotionInput,
) (*model.PromotionCodeDTO, error) {
	if input == nil {
		return nil, errors.New("input is required")
	}
	if id <= 0 {
		return nil, errors.New("invalid id")
	}

	tmp := &model.CreatePromotionInput{
		Scopes:     input.Scopes,
		Conditions: input.Conditions,
	}

	if err := validatePromotionInput(tmp); err != nil {
		return nil, err
	}

	item, err := s.repo.UpdatePromotion(ctx, id, input)
	if err != nil {
		return nil, err
	}
	return item, nil
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
) (*model.PromotionCodeDTO, error) {
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
