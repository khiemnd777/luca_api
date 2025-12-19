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

	GetProcessByID(
		ctx context.Context,
		inProgressID int64,
	) (*model.OrderItemProcessInProgressAndProcessDTO, error)

	GetProcessesByProcessID(
		ctx context.Context,
		processID int64,
	) ([]*model.OrderItemProcessInProgressAndProcessDTO, error)

	PrepareCheckInOrOut(
		ctx context.Context,
		orderID int64,
		orderItemID int64,
	) (*model.OrderItemProcessInProgressDTO, error)

	PrepareCheckInOrOutByCode(
		ctx context.Context,
		code string,
	) (*model.OrderItemProcessInProgressDTO, error)

	CheckInOrOut(
		ctx context.Context,
		checkInOrOutData *model.OrderItemProcessInProgressDTO,
		note *string,
	) (*model.OrderItemProcessInProgressDTO, error)

	Update(
		ctx context.Context,
		deptID int,
		input *model.OrderItemProcessUpsertDTO,
	) (*model.OrderItemProcessDTO, error)
}

type orderItemProcessService struct {
	repo           repository.OrderItemProcessRepository
	inprogressRepo repository.OrderItemProcessInProgressRepository
	deps           *module.ModuleDeps[config.ModuleConfig]
	cfMgr          *customfields.Manager
}

func NewOrderItemProcessService(
	repo repository.OrderItemProcessRepository,
	inprogressRepo repository.OrderItemProcessInProgressRepository,
	deps *module.ModuleDeps[config.ModuleConfig],
	cfMgr *customfields.Manager,
) OrderItemProcessService {
	return &orderItemProcessService{
		repo:           repo,
		inprogressRepo: inprogressRepo,
		deps:           deps,
		cfMgr:          cfMgr,
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

func (s *orderItemProcessService) GetProcessByID(ctx context.Context, inProgressID int64) (*model.OrderItemProcessInProgressAndProcessDTO, error) {
	return cache.Get(fmt.Sprintf("order:process:inprogress:id%d", inProgressID), cache.TTLMedium, func() (*model.OrderItemProcessInProgressAndProcessDTO, error) {
		return s.inprogressRepo.GetProcessByID(ctx, nil, inProgressID)
	})
}

func (s *orderItemProcessService) GetProcessesByProcessID(ctx context.Context, processID int64) ([]*model.OrderItemProcessInProgressAndProcessDTO, error) {
	return cache.GetList(fmt.Sprintf("order:process:id%d:inprogresses", processID), cache.TTLMedium, func() ([]*model.OrderItemProcessInProgressAndProcessDTO, error) {
		return s.inprogressRepo.GetProcessesByProcessID(ctx, nil, processID)
	})
}

func (s *orderItemProcessService) PrepareCheckInOrOut(ctx context.Context, orderID int64, orderItemID int64) (*model.OrderItemProcessInProgressDTO, error) {
	return s.inprogressRepo.PrepareCheckInOrOut(ctx, orderItemID, &orderID)
}

func (s *orderItemProcessService) PrepareCheckInOrOutByCode(ctx context.Context, code string) (*model.OrderItemProcessInProgressDTO, error) {
	return s.inprogressRepo.PrepareCheckInOrOutByCode(ctx, code)
}

// TODO: remove all orderID, orderItemID
func (s *orderItemProcessService) CheckInOrOut(ctx context.Context, checkInOrOutData *model.OrderItemProcessInProgressDTO, note *string) (*model.OrderItemProcessInProgressDTO, error) {
	var err error
	dto, err := s.inprogressRepo.CheckInOrOut(ctx, checkInOrOutData, note)
	if err != nil {
		return nil, err
	}

	orderItemID := dto.OrderItemID
	if orderItemID == 0 && checkInOrOutData != nil {
		orderItemID = checkInOrOutData.OrderItemID
	}
	orderID := dto.OrderID
	if orderID == nil && checkInOrOutData != nil {
		orderID = checkInOrOutData.OrderID
	}

	var keys []string
	if orderID != nil {
		keys = append(keys, kOrderByID(*orderID), kOrderByIDAll(*orderID))
		if orderItemID > 0 {
			keys = append(keys, fmt.Sprintf("order:id:%d:oid:%d:processes", *orderID, orderItemID))
		}
	}

	keys = append(keys, fmt.Sprintf("order:process:inprogress:id%d", dto.ID))
	if dto.ProcessID != nil {
		keys = append(keys, fmt.Sprintf("order:process:id%d:inprogresses", *dto.ProcessID))
	}
	if dto.PrevProcessID != nil {
		keys = append(keys, fmt.Sprintf("order:process:id%d:inprogresses", *dto.PrevProcessID))
	}
	if dto.NextProcessID != nil {
		keys = append(keys, fmt.Sprintf("order:process:id%d:inprogresses", *dto.NextProcessID))
	}

	cache.InvalidateKeys(keys...)
	cache.InvalidateKeys(kOrderAll()...)

	return dto, nil
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
