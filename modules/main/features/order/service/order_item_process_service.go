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

type OrderItemProcessService interface {
}

type orderItemProcessService struct {
	repo  repository.OrderItemProcessRepository
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewOrderItemProcessService(
	repo repository.OrderItemProcessRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
	cfMgr *customfields.Manager,
) OrderItemProcessService {
	return &orderItemProcessService{
		repo:  repo,
		deps:  deps,
		cfMgr: cfMgr,
	}
}

func (s *orderItemProcessService) GetRawProcessesByProductID(ctx context.Context, productID int) ([]*model.ProcessDTO, error) {
	return cache.GetList(fmt.Sprintf("product:id:%d:processes", productID), cache.TTLMedium, func() ([]*model.ProcessDTO, error) {
		return s.repo.GetRawProcessesByProductID(ctx, productID)
	})
}
