package service

import (
	"context"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/order/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/modules/realtime"
	searchmodel "github.com/khiemnd777/andy_api/shared/modules/search/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
)

type OrderItemService interface {
	CalculateTotalPrice(prices []float64, quantities []int) float64

	SyncPrice(ctx context.Context, orderItemID int64) (float64, error)

	GetAllProductsAndMaterialsByOrderID(ctx context.Context, orderID int64) (model.OrderProductsAndMaterialsDTO, error)

	GetHistoricalByOrderIDAndOrderItemID(
		ctx context.Context,
		orderID, orderItemID int64,
	) ([]*model.OrderItemHistoricalDTO, error)

	GetLatestOrderItemIDByOrderID(ctx context.Context, orderID int64) (int64, error)
	GetOrderIDAndOrderItemIDByCode(ctx context.Context, code string) (int64, int64, error)
	Delete(ctx context.Context, deptID int, orderID, orderItemID int64) error
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

func (s *orderItemService) GetLatestOrderItemIDByOrderID(ctx context.Context, orderID int64) (int64, error) {
	return s.repo.GetLatestOrderItemIDByOrderID(ctx, orderID)
}

func (s *orderItemService) GetOrderIDAndOrderItemIDByCode(ctx context.Context, code string) (int64, int64, error) {
	return s.repo.GetOrderIDAndOrderItemIDByCode(ctx, code)
}

func (s *orderItemService) Delete(ctx context.Context, deptID int, orderID, orderItemID int64) error {
	err := s.repo.Delete(ctx, orderItemID)
	if err != nil {
		return err
	}

	cache.InvalidateKeys(
		fmt.Sprintf("order:id:%d:historical:oid:%d", orderID, orderItemID),
		fmt.Sprintf("order:id:%d:historical:oid:0", orderID),
		"order:list:*",
		"order:search:*",
	)

	pubsub.PublishAsync("dashboard:daily:active:stats", &model.CaseDailyActiveStatsUpsert{
		DepartmentID: deptID,
		StatAt:       time.Now(),
	})

	pubsub.PublishAsync("dashboard:daily:sales", &model.SalesDailyUpsert{
		DepartmentID: deptID,
		StatAt:       time.Now(),
	})

	realtime.BroadcastToDept(deptID, "dashboard:daily:active:stats", nil)
	realtime.BroadcastToDept(deptID, "dashboard:statuses", nil)
	realtime.BroadcastToDept(deptID, "dashboard:due_today", nil)
	realtime.BroadcastToDept(deptID, "dashboard:active_today", nil)
	realtime.BroadcastToDept(deptID, "dashboard:sales_summary", nil)
	realtime.BroadcastToDept(deptID, "dashboard:sales_daily", nil)

	s.unlinkSearch(orderID)

	return nil
}

func (s *orderItemService) unlinkSearch(id int64) {
	pubsub.PublishAsync("search:unlink", &searchmodel.UnlinkDoc{
		EntityType: "order",
		EntityID:   id,
	})
}
