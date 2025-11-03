package service

import (
	// "context"

	"fmt"

	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"

	"context"

	"github.com/khiemnd777/andy_api/modules/attribute/config"
	"github.com/khiemnd777/andy_api/modules/attribute/repository"
	"github.com/khiemnd777/andy_api/shared/module"
)

type AttributeService struct {
	repo *repository.AttributeRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewAttributeService(repo *repository.AttributeRepository, deps *module.ModuleDeps[config.ModuleConfig]) *AttributeService {
	return &AttributeService{
		repo: repo,
		deps: deps,
	}
}

func (s *AttributeService) CreateAttribute(ctx context.Context, userId int, name, attributeType string) (*generated.Attribute, error) {
	key := fmt.Sprintf("user:%d:attribute:list", userId)
	keyListFirstPage := fmt.Sprintf("user:%d:attribute:list:first-page", userId)

	var attrResult *generated.Attribute
	err := cache.UpdateManyAndInvalidate([]string{key, keyListFirstPage}, func() error {
		attr, err := s.repo.Create(ctx, &generated.Attribute{
			UserID:        userId,
			AttributeName: name,
			AttributeType: attributeType,
		})
		attrResult = attr
		return err
	})
	return attrResult, err
}

func (s *AttributeService) GetAttribute(ctx context.Context, id, userId int) (*generated.Attribute, error) {
	key := fmt.Sprintf("user:%d:attribute:id:%d", userId, id)
	return cache.Get(key, cache.TTLShort, func() (*generated.Attribute, error) {
		return s.repo.GetByID(ctx, id)
	})
}

func (s *AttributeService) ListAttributes(ctx context.Context, userId int) ([]*generated.Attribute, error) {
	key := fmt.Sprintf("user:%d:attribute:list", userId)
	return cache.GetList(key, cache.TTLLong, func() ([]*generated.Attribute, error) {
		return s.repo.ListByUser(ctx, userId)
	})
}

func (s *AttributeService) ListAttributesPaginated(ctx context.Context, userId, page, limit int) ([]*generated.Attribute, bool, error) {
	offset := (page - 1) * limit
	if page == 1 {
		key := fmt.Sprintf("user:%d:attribute:list:first-page", userId)
		return cache.GetListWithHasMore(key, cache.TTLLong, func() ([]*generated.Attribute, bool, error) {
			return s.repo.ListByUserPaginated(ctx, userId, limit, offset)
		})
	}
	return s.repo.ListByUserPaginated(ctx, userId, limit, offset)
}

func (s *AttributeService) UpdateAttribute(ctx context.Context, id, userId int, name, attributeType string) (*generated.Attribute, error) {
	keyAttrId := fmt.Sprintf("user:%d:attribute:id:%d", userId, id)
	keyUsrId := fmt.Sprintf("user:%d:attribute:list", userId)
	keyListFirstPage := fmt.Sprintf("user:%d:attribute:list:first-page", userId)

	cache.UpdateManyAndInvalidate([]string{keyListFirstPage, keyAttrId, keyUsrId}, func() error {
		_, err := s.repo.Update(ctx, id, name, attributeType)
		return err
	})

	return s.GetAttribute(ctx, id, userId)
}

func (s *AttributeService) DeleteAttribute(ctx context.Context, id, userId int) error {
	keyAttrId := fmt.Sprintf("user:%d:attribute:id:%d", userId, id)
	keyUsrId := fmt.Sprintf("user:%d:attribute:list", userId)
	keyListFirstPage := fmt.Sprintf("user:%d:attribute:list:first-page", userId)

	return cache.UpdateManyAndInvalidate([]string{keyListFirstPage, keyAttrId, keyUsrId}, func() error {
		return s.repo.SoftDelete(ctx, id)
	})
}
