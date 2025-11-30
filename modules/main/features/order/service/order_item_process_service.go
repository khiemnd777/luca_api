package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/order/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
)

type OrderItemProcessService interface {
	GetRawProcessesByProductID(ctx context.Context, productID int) ([]*model.ProcessDTO, error)

	GetProcessesByOrderItemID(
		ctx context.Context,
		orderID int64,
		orderItemID int64,
	) ([]*model.OrderItemProcessDTO, error)

	GetProcessesByAssignedID(
		ctx context.Context,
		assignedID int64,
	) ([]*model.OrderItemProcessDTO, error)

	Update(
		ctx context.Context,
		deptID int,
		input *model.OrderItemProcessUpsertDTO,
	) (*model.OrderItemProcessDTO, error)
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

func (s *orderItemProcessService) GetProcessesByOrderItemID(
	ctx context.Context,
	orderID int64,
	orderItemID int64,
) ([]*model.OrderItemProcessDTO, error) {
	return cache.GetList(fmt.Sprintf("order:id:%d:oid:%d:processes", orderID, orderItemID), cache.TTLShort, func() ([]*model.OrderItemProcessDTO, error) {
		return s.repo.GetProcessesByOrderItemID(ctx, nil, orderItemID)
	})
}

func (s *orderItemProcessService) GetProcessesByAssignedID(
	ctx context.Context,
	assignedID int64,
) ([]*model.OrderItemProcessDTO, error) {
	return cache.GetList(fmt.Sprintf("order:assigned:%d:processes", assignedID), cache.TTLShort, func() ([]*model.OrderItemProcessDTO, error) {
		return s.repo.GetProcessesByAssignedID(ctx, nil, assignedID)
	})
}

func (s *orderItemProcessService) Update(ctx context.Context, deptID int, input *model.OrderItemProcessUpsertDTO) (*model.OrderItemProcessDTO, error) {
	var err error
	tx, err := s.deps.Ent.(*generated.Client).Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	dto, err := s.repo.Update(ctx, tx, input.DTO.ID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(
			kOrderByID(dto.ID),
			kOrderByIDAll(dto.ID),
			fmt.Sprintf("order:id:%d:oid:%d:processes", *dto.OrderID, dto.OrderItemID),
		)
	}
	cache.InvalidateKeys(kOrderAll()...)

	return dto, nil
}
