package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/attribute/config"
	attribute "github.com/khiemnd777/andy_api/modules/attribute/model"
	"github.com/khiemnd777/andy_api/modules/attribute/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/module"
)

type AttributeOptionService struct {
	repo *repository.AttributeOptionRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewAttributeOptionService(repo *repository.AttributeOptionRepository, deps *module.ModuleDeps[config.ModuleConfig]) *AttributeOptionService {
	return &AttributeOptionService{
		repo: repo,
		deps: deps,
	}
}

func (s *AttributeOptionService) ListOptions(ctx context.Context, userID, attributeID int) ([]*generated.AttributeOption, error) {
	key := fmt.Sprintf("user:%d:attribute:%d:options:list", userID, attributeID)
	return cache.GetList(key, cache.TTLLong, func() ([]*generated.AttributeOption, error) {
		return s.repo.ListByAttribute(ctx, attributeID)
	})
}

func (s *AttributeOptionService) CreateOption(ctx context.Context, userID, attributeID int, value string, order int) (*generated.AttributeOption, error) {
	var result *generated.AttributeOption
	key := fmt.Sprintf("user:%d:attribute:%d:options:list", userID, attributeID)

	err := cache.UpdateManyAndInvalidate([]string{key}, func() error {
		opt, err := s.repo.Create(ctx, &generated.AttributeOption{
			UserID:       userID,
			AttributeID:  attributeID,
			OptionValue:  value,
			DisplayOrder: order,
		})
		result = opt
		return err
	})

	return result, err
}

func (s *AttributeOptionService) UpdateOption(ctx context.Context, userID, attributeID, optionID int, value string, order int) (*generated.AttributeOption, error) {
	keyList := fmt.Sprintf("user:%d:attribute:%d:options:list", userID, attributeID)
	keySingle := fmt.Sprintf("user:%d:attribute:%d:options:%d", userID, attributeID, optionID)

	var result *generated.AttributeOption
	err := cache.UpdateManyAndInvalidate([]string{keyList, keySingle}, func() error {
		attrOpt, err := s.repo.Update(ctx, optionID, value, order)
		result = attrOpt
		return err
	})
	return result, err
}

func (s *AttributeOptionService) GetOption(ctx context.Context, userID, attributeID, optionID int) (*generated.AttributeOption, error) {
	key := fmt.Sprintf("user:%d:attribute:%d:options:%d", userID, attributeID, optionID)
	return cache.Get(key, cache.TTLShort, func() (*generated.AttributeOption, error) {
		return s.repo.GetByID(ctx, optionID)
	})
}

func (s *AttributeOptionService) DeleteOption(ctx context.Context, userID, attributeID, optionID int) error {
	keyList := fmt.Sprintf("user:%d:attribute:%d:options:list", userID, attributeID)
	keySingle := fmt.Sprintf("user:%d:attribute:%d:options:%d", userID, attributeID, optionID)

	return cache.UpdateManyAndInvalidate([]string{keyList, keySingle}, func() error {
		return s.repo.SoftDelete(ctx, optionID)
	})
}

func (s *AttributeOptionService) BatchUpdateDisplayOrder(
	ctx context.Context,
	userID, attributeID int,
	orders []attribute.OptionOrder,
) error {
	keyList := fmt.Sprintf("user:%d:attribute:%d:options:list", userID, attributeID)

	return cache.UpdateManyAndInvalidate([]string{keyList}, func() error {
		return s.repo.BatchUpdateDisplayOrder(ctx, orders)
	})
}
