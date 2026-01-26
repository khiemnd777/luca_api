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
	"github.com/khiemnd777/andy_api/shared/modules/notification"
	"github.com/khiemnd777/andy_api/shared/modules/realtime"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	"github.com/khiemnd777/andy_api/shared/utils/table"
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

	GetInProgressByID(
		ctx context.Context,
		inProgressID int64,
	) (*model.OrderItemProcessInProgressAndProcessDTO, error)

	GetInProgressesByProcessID(
		ctx context.Context,
		processID int64,
	) ([]*model.OrderItemProcessInProgressAndProcessDTO, error)

	GetInProgressesByOrderItemID(
		ctx context.Context,
		orderID int64,
		orderItemID int64,
	) ([]*model.OrderItemProcessInProgressAndProcessDTO, error)

	GetInProgressesByAssignedID(
		ctx context.Context,
		assignedID int64,
		query table.TableQuery,
	) (table.TableListResult[model.OrderItemProcessInProgressAndProcessDTO], error)
	GetCheckoutLatest(
		ctx context.Context,
		orderItemID int64,
	) (*model.OrderItemProcessInProgressAndProcessDTO, error)

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
		deptID int,
		userID int,
		checkInOrOutData *model.OrderItemProcessInProgressDTO,
	) (*model.OrderItemProcessInProgressDTO, error)
	Assign(
		ctx context.Context,
		inprogressID int64,
		assignedID *int64,
		assignedName *string,
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

func kAssignedInProgressList(assignedID int64, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("order:assigned:%d:inprogresses:l%d:p%d:o%s:d%s", assignedID, q.Limit, q.Page, orderBy, q.Direction)
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

func (s *orderItemProcessService) GetInProgressByID(ctx context.Context, inProgressID int64) (*model.OrderItemProcessInProgressAndProcessDTO, error) {
	return cache.Get(fmt.Sprintf("order:process:inprogress:id%d", inProgressID), cache.TTLMedium, func() (*model.OrderItemProcessInProgressAndProcessDTO, error) {
		return s.inprogressRepo.GetInProgressByID(ctx, nil, inProgressID)
	})
}

func (s *orderItemProcessService) GetInProgressesByProcessID(ctx context.Context, processID int64) ([]*model.OrderItemProcessInProgressAndProcessDTO, error) {
	return cache.GetList(fmt.Sprintf("order:process:id%d:inprogresses", processID), cache.TTLMedium, func() ([]*model.OrderItemProcessInProgressAndProcessDTO, error) {
		return s.inprogressRepo.GetInProgressesByProcessID(ctx, nil, processID)
	})
}

func (s *orderItemProcessService) GetInProgressesByOrderItemID(
	ctx context.Context,
	orderID int64,
	orderItemID int64,
) ([]*model.OrderItemProcessInProgressAndProcessDTO, error) {
	return cache.GetList(fmt.Sprintf("order:id:%d:oid:%d:inprogresses", orderID, orderItemID), cache.TTLShort, func() ([]*model.OrderItemProcessInProgressAndProcessDTO, error) {
		return s.inprogressRepo.GetInProgressesByOrderItemID(ctx, nil, orderItemID)
	})
}

func (s *orderItemProcessService) GetInProgressesByAssignedID(
	ctx context.Context,
	assignedID int64,
	query table.TableQuery,
) (table.TableListResult[model.OrderItemProcessInProgressAndProcessDTO], error) {
	type boxed = table.TableListResult[model.OrderItemProcessInProgressAndProcessDTO]
	key := kAssignedInProgressList(assignedID, query)

	ptr, err := cache.Get(key, cache.TTLShort, func() (*boxed, error) {
		res, e := s.inprogressRepo.GetInProgressesByAssignedID(ctx, nil, assignedID, query)
		if e != nil {
			return nil, e
		}
		return &res, nil
	})
	if err != nil {
		var zero boxed
		return zero, err
	}
	return *ptr, nil
}

func (s *orderItemProcessService) GetCheckoutLatest(ctx context.Context, orderItemID int64) (*model.OrderItemProcessInProgressAndProcessDTO, error) {
	return cache.Get(fmt.Sprintf("order:process:checkout:latest:oid:%d", orderItemID), cache.TTLShort, func() (*model.OrderItemProcessInProgressAndProcessDTO, error) {
		return s.inprogressRepo.GetCheckoutLatest(ctx, nil, orderItemID)
	})
}

func (s *orderItemProcessService) PrepareCheckInOrOut(ctx context.Context, orderID int64, orderItemID int64) (*model.OrderItemProcessInProgressDTO, error) {
	return s.inprogressRepo.PrepareCheckInOrOut(ctx, nil, orderItemID, &orderID)
}

func (s *orderItemProcessService) PrepareCheckInOrOutByCode(ctx context.Context, code string) (*model.OrderItemProcessInProgressDTO, error) {
	return s.inprogressRepo.PrepareCheckInOrOutByCode(ctx, code)
}

// TODO: remove all orderID, orderItemID
func (s *orderItemProcessService) CheckInOrOut(
	ctx context.Context,
	deptID,
	userID int,
	checkInOrOutData *model.OrderItemProcessInProgressDTO,
) (*model.OrderItemProcessInProgressDTO, error) {
	var err error
	dto, _, orderstatus, orderitem, err := s.inprogressRepo.CheckInOrOut(ctx, checkInOrOutData)
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
			keys = append(keys, fmt.Sprintf("order:id:%d:oid:%d:inprogresses", *orderID, orderItemID))
		}
	}

	if orderItemID > 0 {
		keys = append(keys, fmt.Sprintf("order:process:checkout:latest:oid:%d", orderItemID))
	}

	keys = append(keys, fmt.Sprintf("order:process:inprogress:id%d", dto.ID))
	if dto.ProcessID != nil {
		keys = append(keys, fmt.Sprintf("order:process:id%d:*", *dto.ProcessID))
	}
	if dto.PrevProcessID != nil {
		keys = append(keys, fmt.Sprintf("order:process:id%d:*", *dto.PrevProcessID))
	}
	if dto.NextProcessID != nil {
		keys = append(keys, fmt.Sprintf("order:process:id%d:*", *dto.NextProcessID))
	}
	if dto.AssignedID != nil {
		keys = append(keys, fmt.Sprintf("order:assigned:%d:*", *dto.AssignedID))
	}

	cache.InvalidateKeys(keys...)
	cache.InvalidateKeys(kOrderAll()...)

	// notify to next process's leader
	if dto.CompletedAt != nil && dto.NextProcessID != nil {
		notification.Notify(*dto.NextLeaderID, userID, "order:checkout", map[string]any{
			"leader_id":       dto.NextLeaderID,
			"leader_name":     dto.NextLeaderName,
			"order_item_id":   dto.OrderItemID,
			"order_item_code": dto.OrderItemCode,
			"section_name":    dto.NextSectionName,
			"process_name":    dto.NextProcessName,
		})

		if orderstatus != nil && "completed" == *orderstatus {
			pubsub.PublishAsync("dashboard:daily:stats", &model.CaseDailyStatsUpsert{
				DepartmentID: deptID,
				CompletedAt:  *dto.CompletedAt,
				ReceivedAt:   orderitem.CreatedAt,
			})

			pubsub.PublishAsync("dashboard:daily:remake:stats", &model.CaseDailyRemakeStatsUpsert{
				DepartmentID: deptID,
				CompletedAt:  *dto.CompletedAt,
				IsRemake:     orderitem.RemakeCount > 0,
			})
		}
	}

	realtime.BroadcastAll("order:inprogress", nil)

	return dto, nil
}

func (s *orderItemProcessService) Assign(ctx context.Context, inprogressID int64, assignedID *int64, assignedName *string, note *string) (*model.OrderItemProcessInProgressDTO, error) {
	dto, _, _, _, err := s.inprogressRepo.Assign(ctx, inprogressID, assignedID, assignedName, note)
	if err != nil {
		return nil, err
	}

	var keys []string
	if dto.OrderID != nil {
		keys = append(keys, kOrderByID(*dto.OrderID), kOrderByIDAll(*dto.OrderID))
		if dto.OrderItemID > 0 {
			keys = append(keys, fmt.Sprintf("order:id:%d:oid:%d:processes", *dto.OrderID, dto.OrderItemID))
			keys = append(keys, fmt.Sprintf("order:id:%d:oid:%d:inprogresses", *dto.OrderID, dto.OrderItemID))
		}
	}

	if dto.OrderItemID > 0 {
		keys = append(keys, fmt.Sprintf("order:process:checkout:latest:oid:%d", dto.OrderItemID))
	}

	keys = append(keys, fmt.Sprintf("order:process:inprogress:id%d", dto.ID))
	if dto.ProcessID != nil {
		keys = append(keys, fmt.Sprintf("order:process:id%d:*", *dto.ProcessID))
	}
	if dto.PrevProcessID != nil {
		keys = append(keys, fmt.Sprintf("order:process:id%d:*", *dto.PrevProcessID))
	}
	if dto.NextProcessID != nil {
		keys = append(keys, fmt.Sprintf("order:process:id%d:*", *dto.NextProcessID))
	}
	if dto.AssignedID != nil {
		keys = append(keys, fmt.Sprintf("order:assigned:%d:*", *dto.AssignedID))
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
