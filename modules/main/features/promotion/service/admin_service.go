package service

import (
	"context"
	"errors"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/promotion/repository"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type PromotionAdminService interface {
	CreatePromotion(ctx context.Context, input *model.CreatePromotionInput) (*model.PromotionCodeDTO, error)
	UpdatePromotion(ctx context.Context, id int, input *model.UpdatePromotionInput) (*model.PromotionCodeDTO, error)
	DeletePromotion(ctx context.Context, id int) error
	GetPromotionByID(ctx context.Context, id int) (*model.PromotionCodeDTO, error)
	ListPromotions(ctx context.Context, query table.TableQuery) (table.TableListResult[model.PromotionCodeDTO], error)
}

type promotionAdminService struct {
	repo repository.PromotionAdminRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewPromotionAdminService(repo repository.PromotionAdminRepository, deps *module.ModuleDeps[config.ModuleConfig]) PromotionAdminService {
	return &promotionAdminService{
		repo: repo,
		deps: deps,
	}
}

func (s *promotionAdminService) CreatePromotion(
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

func (s *promotionAdminService) UpdatePromotion(
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

func (s *promotionAdminService) DeletePromotion(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return s.repo.DeletePromotion(ctx, id)
}

func (s *promotionAdminService) GetPromotionByID(
	ctx context.Context,
	id int,
) (*model.PromotionCodeDTO, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.repo.GetPromotionByID(ctx, id)
}

func (s *promotionAdminService) ListPromotions(
	ctx context.Context,
	query table.TableQuery,
) (table.TableListResult[model.PromotionCodeDTO], error) {
	return s.repo.ListPromotions(ctx, query)
}
