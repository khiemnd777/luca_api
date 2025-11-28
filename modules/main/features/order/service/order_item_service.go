package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/order/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
)

type OrderItemService interface {
	GetHistoricalByOrderIDAndOrderItemID(
		ctx context.Context,
		orderID, orderItemID int64,
	) ([]*model.OrderItemHistoricalDTO, error)
}

type orderItemService struct {
	repo  repository.OrderItemRepository
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewOrderItemService(
	repo repository.OrderItemRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
	cfMgr *customfields.Manager,
) OrderItemService {
	return &orderItemService{
		repo:  repo,
		deps:  deps,
		cfMgr: cfMgr,
	}
}

func (s *orderItemService) GetHistoricalByOrderIDAndOrderItemID(
	ctx context.Context,
	orderID, orderItemID int64,
) ([]*model.OrderItemHistoricalDTO, error) {
	return cache.GetList(fmt.Sprintf("order:id:%d:historical:oid:%d", orderID, orderItemID), cache.TTLMedium, func() ([]*model.OrderItemHistoricalDTO, error) {
		return s.repo.GetHistoricalByOrderIDAndOrderItemID(ctx, orderID, orderItemID)
	})
}
