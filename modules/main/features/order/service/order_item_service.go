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
	CalculateTotalPrice(prices []float64, quantities []int) float64

	SyncPrice(ctx context.Context, orderItemID int64) (float64, error)

	GetAllProductsAndMaterialsByOrderID(ctx context.Context, orderID int64) (model.OrderProductsAndMaterialsDTO, error)

	GetHistoricalByOrderIDAndOrderItemID(
		ctx context.Context,
		orderID, orderItemID int64,
	) ([]*model.OrderItemHistoricalDTO, error)

	GetOrderIDAndOrderItemIDByCode(ctx context.Context, code string) (int64, int64, error)
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

func (s *orderItemService) CalculateTotalPrice(prices []float64, quantities []int) float64 {
	var total float64

	for i, price := range prices {
		qty := 1
		if i < len(quantities) && quantities[i] > 0 {
			qty = quantities[i]
		}
		total += price * float64(qty)
	}
	return total
}

func (s *orderItemService) SyncPrice(ctx context.Context, orderItemID int64) (float64, error) {
	return s.repo.GetTotalPriceByOrderItemID(ctx, orderItemID)
}

func (s *orderItemService) GetAllProductsAndMaterialsByOrderID(ctx context.Context, orderID int64) (model.OrderProductsAndMaterialsDTO, error) {
	return s.repo.GetAllProductsAndMaterialsByOrderID(ctx, orderID)
}

func (s *orderItemService) GetHistoricalByOrderIDAndOrderItemID(
	ctx context.Context,
	orderID, orderItemID int64,
) ([]*model.OrderItemHistoricalDTO, error) {
	return cache.GetList(fmt.Sprintf("order:id:%d:historical:oid:%d", orderID, orderItemID), cache.TTLMedium, func() ([]*model.OrderItemHistoricalDTO, error) {
		return s.repo.GetHistoricalByOrderIDAndOrderItemID(ctx, orderID, orderItemID)
	})
}

func (s *orderItemService) GetOrderIDAndOrderItemIDByCode(ctx context.Context, code string) (int64, int64, error) {
	return s.repo.GetOrderIDAndOrderItemIDByCode(ctx, code)
}
