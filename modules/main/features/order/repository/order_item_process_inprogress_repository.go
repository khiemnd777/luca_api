package repository

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitem"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemprocess"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemprocessinprogress"
	"github.com/khiemnd777/andy_api/shared/mapper"
)

type OrderItemProcessInProgressRepository interface {
	PrepareCheckInOrOut(ctx context.Context, orderItemID int64, orderID *int64) (*model.OrderItemProcessInProgressDTO, error)
	PrepareCheckInOrOutByCode(ctx context.Context, code string) (*model.OrderItemProcessInProgressDTO, error)
	CheckInOrOut(ctx context.Context, checkInOrOutData *model.OrderItemProcessInProgressDTO, note *string) (*model.OrderItemProcessInProgressDTO, error)
	Assign(ctx context.Context, inprogressID int64, assignedID *int64, assignedName *string, note *string) (*model.OrderItemProcessInProgressDTO, error)
	CheckIn(ctx context.Context, tx *generated.Tx, orderItemID int64, orderID *int64, note *string) (*model.OrderItemProcessInProgressDTO, error)
	CheckOut(ctx context.Context, tx *generated.Tx, orderItemID int64, note *string) (*model.OrderItemProcessInProgressDTO, error)
	GetLatest(ctx context.Context, tx *generated.Tx, orderItemID int64) (*model.OrderItemProcessInProgressDTO, error)
	GetCheckoutLatest(ctx context.Context, tx *generated.Tx, orderItemID int64) (*model.OrderItemProcessInProgressAndProcessDTO, error)
	GetInProgressesByOrderItemID(ctx context.Context, tx *generated.Tx, orderItemID int64) ([]*model.OrderItemProcessInProgressAndProcessDTO, error)
	GetInProgressesByProcessID(ctx context.Context, tx *generated.Tx, processID int64) ([]*model.OrderItemProcessInProgressAndProcessDTO, error)
	GetInProgressByID(ctx context.Context, tx *generated.Tx, inProgressID int64) (*model.OrderItemProcessInProgressAndProcessDTO, error)
}

type orderItemProcessInProgressRepository struct {
	db                   *generated.Client
	orderItemProcessRepo OrderItemProcessRepository
}

func NewOrderItemProcessInProgressRepository(db *generated.Client, orderItemProcessRepo OrderItemProcessRepository) OrderItemProcessInProgressRepository {
	return &orderItemProcessInProgressRepository{db: db, orderItemProcessRepo: orderItemProcessRepo}
}

func (r *orderItemProcessInProgressRepository) GetInProgressesByProcessID(ctx context.Context, tx *generated.Tx, processID int64) ([]*model.OrderItemProcessInProgressAndProcessDTO, error) {
	items, err := r.inprogressClient(tx).
		Query().
		Where(orderitemprocessinprogress.ProcessID(processID)).
		Order(orderitemprocessinprogress.ByCreatedAt(sql.OrderDesc())).
		Select(
			orderitemprocessinprogress.FieldID,
			orderitemprocessinprogress.FieldNote,
			orderitemprocessinprogress.FieldAssignedID,
			orderitemprocessinprogress.FieldAssignedName,
			orderitemprocessinprogress.FieldStartedAt,
			orderitemprocessinprogress.FieldCompletedAt,
		).
		WithProcess(func(q *generated.OrderItemProcessQuery) {
			q.Select(
				orderitemprocess.FieldID,
				orderitemprocess.FieldProcessName,
				orderitemprocess.FieldSectionName,
				orderitemprocess.FieldColor,
			)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*model.OrderItemProcessInProgressAndProcessDTO, 0, len(items))
	for _, item := range items {
		proc, err := item.Edges.ProcessOrErr()
		if err != nil {
			return nil, err
		}
		out = append(out, &model.OrderItemProcessInProgressAndProcessDTO{
			ID:           item.ID,
			Note:         item.Note,
			AssignedID:   item.AssignedID,
			AssignedName: item.AssignedName,
			StartedAt:    item.StartedAt,
			CompletedAt:  item.CompletedAt,
			ProcessName:  proc.ProcessName,
			SectionName:  proc.SectionName,
			Color:        proc.Color,
		})
	}

	return out, nil
}

func (r *orderItemProcessInProgressRepository) GetInProgressesByOrderItemID(ctx context.Context, tx *generated.Tx, orderItemID int64) ([]*model.OrderItemProcessInProgressAndProcessDTO, error) {
	items, err := r.inprogressClient(tx).
		Query().
		Where(orderitemprocessinprogress.OrderItemID(orderItemID)).
		Order(orderitemprocessinprogress.ByCreatedAt(sql.OrderDesc())).
		Select(
			orderitemprocessinprogress.FieldID,
			orderitemprocessinprogress.FieldNote,
			orderitemprocessinprogress.FieldAssignedID,
			orderitemprocessinprogress.FieldAssignedName,
			orderitemprocessinprogress.FieldStartedAt,
			orderitemprocessinprogress.FieldCompletedAt,
		).
		WithProcess(func(q *generated.OrderItemProcessQuery) {
			q.Select(
				orderitemprocess.FieldID,
				orderitemprocess.FieldProcessName,
				orderitemprocess.FieldSectionName,
				orderitemprocess.FieldColor,
			)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*model.OrderItemProcessInProgressAndProcessDTO, 0, len(items))
	for _, item := range items {
		proc, err := item.Edges.ProcessOrErr()
		if err != nil {
			return nil, err
		}
		out = append(out, &model.OrderItemProcessInProgressAndProcessDTO{
			ID:           item.ID,
			Note:         item.Note,
			AssignedID:   item.AssignedID,
			AssignedName: item.AssignedName,
			StartedAt:    item.StartedAt,
			CompletedAt:  item.CompletedAt,
			ProcessName:  proc.ProcessName,
			SectionName:  proc.SectionName,
			Color:        proc.Color,
		})
	}

	return out, nil
}

func (r *orderItemProcessInProgressRepository) GetInProgressByID(ctx context.Context, tx *generated.Tx, inProgressID int64) (*model.OrderItemProcessInProgressAndProcessDTO, error) {
	entity, err := r.inprogressClient(tx).
		Query().
		Where(orderitemprocessinprogress.ID(inProgressID)).
		WithProcess(func(q *generated.OrderItemProcessQuery) {
			q.Select(
				orderitemprocess.FieldID,
				orderitemprocess.FieldProcessName,
				orderitemprocess.FieldSectionName,
				orderitemprocess.FieldColor,
			)
		}).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	proc, err := entity.Edges.ProcessOrErr()
	if err != nil {
		return nil, err
	}

	return &model.OrderItemProcessInProgressAndProcessDTO{
		ID:           entity.ID,
		Note:         entity.Note,
		AssignedID:   entity.AssignedID,
		AssignedName: entity.AssignedName,
		StartedAt:    entity.StartedAt,
		CompletedAt:  entity.CompletedAt,
		ProcessName:  proc.ProcessName,
		SectionName:  proc.SectionName,
		Color:        proc.Color,
	}, nil
}

func (r *orderItemProcessInProgressRepository) PrepareCheckInOrOutByCode(ctx context.Context, code string) (*model.OrderItemProcessInProgressDTO, error) {
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}

	var err error
	tx, err := r.db.Tx(ctx)
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

	orderItemEntity, err := tx.OrderItem.
		Query().
		Where(
			orderitem.CodeEQ(code),
			orderitem.DeletedAtIsNil(),
		).
		Select(
			orderitem.FieldID,
			orderitem.FieldOrderID,
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	orderItemID := orderItemEntity.ID
	orderID := orderItemEntity.OrderID

	return r.PrepareCheckInOrOut(ctx, orderItemID, &orderID)
}

func (r *orderItemProcessInProgressRepository) PrepareCheckInOrOut(ctx context.Context, orderItemID int64, orderID *int64) (*model.OrderItemProcessInProgressDTO, error) {
	var err error
	tx, err := r.db.Tx(ctx)
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

	processes, err := r.getProcesses(ctx, tx, orderItemID)
	if err != nil {
		return nil, err
	}
	if len(processes) == 0 {
		err = fmt.Errorf("no processes found for order item %d", orderItemID)
		return nil, err
	}

	latest, err := r.latestEntity(ctx, tx, orderItemID)
	if err != nil && !generated.IsNotFound(err) {
		return nil, err
	}

	// Checkout
	if latest != nil && latest.CompletedAt == nil {
		currentProcessID := processes[0].ID
		if latest.ProcessID != nil {
			currentProcessID = *latest.ProcessID
		}
		nextProcessID := r.nextProcessID(processes, currentProcessID)
		targetProcess := r.findProcess(processes, currentProcessID)
		if targetProcess == nil {
			err = fmt.Errorf("process %d not found for order item %d", currentProcessID, latest.OrderItemID)
			return nil, err
		}

		return &model.OrderItemProcessInProgressDTO{
			ID:            latest.ID,
			ProcessID:     &currentProcessID,
			PrevProcessID: latest.PrevProcessID,
			NextProcessID: nextProcessID,
			OrderItemID:   latest.OrderItemID,
			OrderID:       r.pickOrderID(latest.OrderID, targetProcess),
			AssignedID:    targetProcess.AssignedID,
			AssignedName:  targetProcess.AssignedName,
			SectionName:   targetProcess.SectionName,
			Note:          latest.Note,
			StartedAt:     latest.StartedAt,
			CompletedAt:   latest.CompletedAt,
			UpdatedAt:     latest.UpdatedAt,
		}, nil
	}

	// Checkin
	targetProcessID := processes[0].ID
	var prevProcessID *int64

	if latest != nil {
		switch {
		case latest.NextProcessID != nil:
			targetProcessID = *latest.NextProcessID
			prevProcessID = latest.ProcessID
		case latest.ProcessID != nil:
			targetProcessID = *latest.ProcessID
			prevProcessID = latest.PrevProcessID
		}
	}

	targetProcess := r.findProcess(processes, targetProcessID)
	if targetProcess == nil {
		err = fmt.Errorf("process %d not found for order item %d", targetProcessID, orderItemID)
		return nil, err
	}

	return &model.OrderItemProcessInProgressDTO{
		ProcessID:     &targetProcessID,
		PrevProcessID: prevProcessID,
		OrderItemID:   orderItemID,
		OrderID:       r.pickOrderID(orderID, targetProcess),
		AssignedID:    targetProcess.AssignedID,
		AssignedName:  targetProcess.AssignedName,
		SectionName:   targetProcess.SectionName,
	}, nil
}

func (r *orderItemProcessInProgressRepository) Assign(ctx context.Context, inprogressID int64, assignedID *int64, assignedName *string, note *string) (*model.OrderItemProcessInProgressDTO, error) {
	var err error
	tx, err := r.db.Tx(ctx)
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

	inprogress, err := r.inprogressClient(tx).
		Query().
		Where(orderitemprocessinprogress.ID(inprogressID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	if inprogress.ProcessID == nil {
		err = fmt.Errorf("process id is required for inprogress %d", inprogressID)
		return nil, err
	}

	proc, err := r.processClient(tx).
		Query().
		Where(orderitemprocess.IDEQ(*inprogress.ProcessID)).
		Select(
			orderitemprocess.FieldCustomFields,
			orderitemprocess.FieldStatus,
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	status := proc.Status
	if v, ok := proc.CustomFields["status"]; ok {
		if s, ok := v.(string); ok && s != "" {
			status = s
		}
	}
	if status == "" {
		status = "in_progress"
	}

	entity, err := r.inprogressClient(tx).
		Create().
		SetNillableProcessID(inprogress.ProcessID).
		SetNillablePrevProcessID(inprogress.PrevProcessID).
		SetNillableNextProcessID(inprogress.NextProcessID).
		SetOrderItemID(inprogress.OrderItemID).
		SetNillableOrderID(inprogress.OrderID).
		SetNillableAssignedID(assignedID).
		SetNillableAssignedName(assignedName).
		SetNillableSectionName(inprogress.SectionName).
		SetNillableNote(note).
		SetNillableStartedAt(inprogress.StartedAt).
		SetNillableCompletedAt(inprogress.CompletedAt).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	if err := r.updateProcessStatusAndAssign(
		ctx,
		tx,
		*inprogress.ProcessID,
		status,
		assignedID,
		assignedName,
	); err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.OrderItemProcessInProgress, *model.OrderItemProcessInProgressDTO](entity)
	return dto, nil
}

func (r *orderItemProcessInProgressRepository) CheckInOrOut(ctx context.Context, checkInOrOutData *model.OrderItemProcessInProgressDTO, note *string) (*model.OrderItemProcessInProgressDTO, error) {
	if checkInOrOutData == nil {
		return nil, fmt.Errorf("checkInOrOutData is required")
	}

	var err error
	tx, err := r.db.Tx(ctx)
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

	// Checkout
	if checkInOrOutData.ID > 0 {
		if checkInOrOutData.ProcessID == nil {
			err = fmt.Errorf("process id is required for checkout of order item process %d", checkInOrOutData.ID)
			return nil, err
		}

		completedAt := time.Now()
		entity, err := r.inprogressClient(tx).
			UpdateOneID(checkInOrOutData.ID).
			SetProcessID(*checkInOrOutData.ProcessID).
			SetNillableNextProcessID(checkInOrOutData.NextProcessID).
			SetNillableOrderID(checkInOrOutData.OrderID).
			SetNillableNote(note).
			SetCompletedAt(completedAt).
			Save(ctx)
		if err != nil {
			return nil, err
		}

		if err := r.updateProcessStatus(ctx, tx, *checkInOrOutData.ProcessID, "completed"); err != nil {
			return nil, err
		}

		dto := mapper.MapAs[*generated.OrderItemProcessInProgress, *model.OrderItemProcessInProgressDTO](entity)
		return dto, nil
	}

	if checkInOrOutData.ProcessID == nil {
		err = fmt.Errorf("process id is required for checkin of order item %d", checkInOrOutData.OrderItemID)
		return nil, err
	}

	// Checkin
	startedAt := time.Now()
	entity, err := r.inprogressClient(tx).
		Create().
		SetNillableProcessID(checkInOrOutData.ProcessID).
		SetNillablePrevProcessID(checkInOrOutData.PrevProcessID).
		SetOrderItemID(checkInOrOutData.OrderItemID).
		SetNillableOrderID(checkInOrOutData.OrderID).
		SetNillableAssignedID(checkInOrOutData.AssignedID).
		SetNillableAssignedName(checkInOrOutData.AssignedName).
		SetNillableSectionName(checkInOrOutData.SectionName).
		SetNillableNote(note).
		SetStartedAt(startedAt).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// process's status
	rework, err := r.hasCompletedProcess(ctx, tx, *checkInOrOutData.ProcessID)
	if err != nil {
		return nil, err
	}
	status := "in_progress"
	if rework {
		status = "rework"
	}
	if err := r.updateProcessStatusAndAssign(
		ctx,
		tx,
		*checkInOrOutData.ProcessID,
		status,
		checkInOrOutData.AssignedID,
		checkInOrOutData.AssignedName,
	); err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.OrderItemProcessInProgress, *model.OrderItemProcessInProgressDTO](entity)
	return dto, nil
}

func (r *orderItemProcessInProgressRepository) CheckIn(ctx context.Context, tx *generated.Tx, orderItemID int64, orderID *int64, note *string) (*model.OrderItemProcessInProgressDTO, error) {
	processes, err := r.getProcesses(ctx, tx, orderItemID)
	if err != nil {
		return nil, err
	}
	if len(processes) == 0 {
		return nil, fmt.Errorf("no processes found for order item %d", orderItemID)
	}

	latest, err := r.latestEntity(ctx, tx, orderItemID)
	if err != nil && !generated.IsNotFound(err) {
		return nil, err
	}

	return r.checkinWithData(ctx, tx, latest, processes, orderItemID, orderID, note)
}

func (r *orderItemProcessInProgressRepository) CheckOut(ctx context.Context, tx *generated.Tx, orderItemID int64, note *string) (*model.OrderItemProcessInProgressDTO, error) {
	processes, err := r.getProcesses(ctx, tx, orderItemID)
	if err != nil {
		return nil, err
	}
	if len(processes) == 0 {
		return nil, fmt.Errorf("no processes found for order item %d", orderItemID)
	}

	latest, err := r.latestEntity(ctx, tx, orderItemID)
	if err != nil {
		return nil, err
	}

	return r.checkoutWithData(ctx, tx, latest, processes, note)
}

func (r *orderItemProcessInProgressRepository) GetLatest(ctx context.Context, tx *generated.Tx, orderItemID int64) (*model.OrderItemProcessInProgressDTO, error) {
	entity, err := r.latestEntity(ctx, tx, orderItemID)
	if err != nil {
		return nil, err
	}
	dto := mapper.MapAs[*generated.OrderItemProcessInProgress, *model.OrderItemProcessInProgressDTO](entity)
	return dto, nil
}

func (r *orderItemProcessInProgressRepository) GetCheckoutLatest(ctx context.Context, tx *generated.Tx, orderItemID int64) (*model.OrderItemProcessInProgressAndProcessDTO, error) {
	entity, err := r.inprogressClient(tx).
		Query().
		Where(
			orderitemprocessinprogress.OrderItemID(orderItemID),
			orderitemprocessinprogress.CompletedAtNotNil(),
		).
		Order(orderitemprocessinprogress.ByCreatedAt(sql.OrderDesc())).
		Select(
			orderitemprocessinprogress.FieldID,
			orderitemprocessinprogress.FieldNote,
			orderitemprocessinprogress.FieldAssignedID,
			orderitemprocessinprogress.FieldAssignedName,
			orderitemprocessinprogress.FieldStartedAt,
			orderitemprocessinprogress.FieldCompletedAt,
		).
		WithProcess(func(q *generated.OrderItemProcessQuery) {
			q.Select(
				orderitemprocess.FieldID,
				orderitemprocess.FieldProcessName,
				orderitemprocess.FieldSectionName,
				orderitemprocess.FieldColor,
			)
		}).
		First(ctx)
	if err != nil {
		return nil, err
	}

	proc, err := entity.Edges.ProcessOrErr()
	if err != nil {
		return nil, err
	}

	return &model.OrderItemProcessInProgressAndProcessDTO{
		ID:           entity.ID,
		Note:         entity.Note,
		AssignedID:   entity.AssignedID,
		AssignedName: entity.AssignedName,
		StartedAt:    entity.StartedAt,
		CompletedAt:  entity.CompletedAt,
		ProcessName:  proc.ProcessName,
		SectionName:  proc.SectionName,
		Color:        proc.Color,
	}, nil
}

func (r *orderItemProcessInProgressRepository) latestEntity(ctx context.Context, tx *generated.Tx, orderItemID int64) (*generated.OrderItemProcessInProgress, error) {
	q := r.inprogressClient(tx).
		Query().
		Where(orderitemprocessinprogress.OrderItemID(orderItemID)).
		Order(orderitemprocessinprogress.ByCreatedAt(sql.OrderDesc()))

	entity, err := q.First(ctx)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *orderItemProcessInProgressRepository) getProcesses(ctx context.Context, tx *generated.Tx, orderItemID int64) ([]*generated.OrderItemProcess, error) {
	q := r.processClient(tx).
		Query().
		Where(orderitemprocess.OrderItemID(orderItemID)).
		Order(orderitemprocess.ByStepNumber(sql.OrderAsc()))
	return q.All(ctx)
}

func (r *orderItemProcessInProgressRepository) nextProcessID(processes []*generated.OrderItemProcess, currentID int64) *int64 {
	for i, p := range processes {
		if p.ID == currentID && i+1 < len(processes) {
			nextID := processes[i+1].ID
			return &nextID
		}
	}
	return nil
}

func (r *orderItemProcessInProgressRepository) findProcess(processes []*generated.OrderItemProcess, processID int64) *generated.OrderItemProcess {
	for _, p := range processes {
		if p.ID == processID {
			return p
		}
	}
	return nil
}

func (r *orderItemProcessInProgressRepository) pickOrderID(orderID *int64, proc *generated.OrderItemProcess) *int64 {
	if orderID != nil {
		return orderID
	}
	return proc.OrderID
}

func (r *orderItemProcessInProgressRepository) hasCompletedProcess(ctx context.Context, tx *generated.Tx, processID int64) (bool, error) {
	return r.inprogressClient(tx).
		Query().
		Where(
			orderitemprocessinprogress.ProcessID(processID),
			orderitemprocessinprogress.CompletedAtNotNil(),
		).
		Exist(ctx)
}

func (r *orderItemProcessInProgressRepository) updateProcessStatus(
	ctx context.Context,
	tx *generated.Tx,
	processID int64,
	status string,
) error {
	if r.orderItemProcessRepo == nil {
		return fmt.Errorf("order item process repository is required")
	}

	_, err := r.orderItemProcessRepo.UpdateStatus(
		ctx,
		tx,
		processID,
		status,
	)
	return err
}

func (r *orderItemProcessInProgressRepository) updateProcessStatusAndAssign(
	ctx context.Context,
	tx *generated.Tx,
	processID int64,
	status string,
	assignedID *int64,
	assignedName *string,
) error {
	if r.orderItemProcessRepo == nil {
		return fmt.Errorf("order item process repository is required")
	}

	_, err := r.orderItemProcessRepo.UpdateStatusAndAssign(
		ctx,
		tx,
		processID,
		status,
		assignedID,
		assignedName,
	)
	return err
}

func (r *orderItemProcessInProgressRepository) checkinWithData(ctx context.Context, tx *generated.Tx, latest *generated.OrderItemProcessInProgress, processes []*generated.OrderItemProcess, orderItemID int64, orderID *int64, note *string) (*model.OrderItemProcessInProgressDTO, error) {
	var prevProcessID *int64
	targetProcessID := processes[0].ID

	if latest != nil {
		switch {
		case latest.NextProcessID != nil:
			targetProcessID = *latest.NextProcessID
			prevProcessID = latest.ProcessID
		case latest.ProcessID != nil:
			targetProcessID = *latest.ProcessID
			prevProcessID = latest.PrevProcessID
		}
	}

	targetProcess := r.findProcess(processes, targetProcessID)
	if targetProcess == nil {
		return nil, fmt.Errorf("process %d not found for order item %d", targetProcessID, orderItemID)
	}

	startedAt := time.Now()
	entity, err := r.inprogressClient(tx).
		Create().
		SetNillableProcessID(&targetProcessID).
		SetNillablePrevProcessID(prevProcessID).
		SetOrderItemID(orderItemID).
		SetNillableOrderID(r.pickOrderID(orderID, targetProcess)).
		SetNillableAssignedID(targetProcess.AssignedID).
		SetNillableAssignedName(targetProcess.AssignedName).
		SetNillableNote(note).
		SetStartedAt(startedAt).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.OrderItemProcessInProgress, *model.OrderItemProcessInProgressDTO](entity)
	return dto, nil
}

func (r *orderItemProcessInProgressRepository) checkoutWithData(ctx context.Context, tx *generated.Tx, latest *generated.OrderItemProcessInProgress, processes []*generated.OrderItemProcess, note *string) (*model.OrderItemProcessInProgressDTO, error) {
	currentProcessID := processes[0].ID
	if latest.ProcessID != nil {
		currentProcessID = *latest.ProcessID
	}

	nextProcessID := r.nextProcessID(processes, currentProcessID)
	targetProcess := r.findProcess(processes, currentProcessID)
	if targetProcess == nil {
		return nil, fmt.Errorf("process %d not found for order item %d", currentProcessID, latest.OrderItemID)
	}

	completedAt := time.Now()
	entity, err := r.inprogressClient(tx).
		UpdateOneID(latest.ID).
		SetProcessID(currentProcessID).
		SetNillableNextProcessID(nextProcessID).
		SetNillableOrderID(r.pickOrderID(latest.OrderID, targetProcess)).
		SetNillableNote(note).
		SetCompletedAt(completedAt).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.OrderItemProcessInProgress, *model.OrderItemProcessInProgressDTO](entity)
	return dto, nil
}

func (r *orderItemProcessInProgressRepository) inprogressClient(tx *generated.Tx) *generated.OrderItemProcessInProgressClient {
	if tx != nil {
		return tx.OrderItemProcessInProgress
	}
	return r.db.OrderItemProcessInProgress
}

func (r *orderItemProcessInProgressRepository) processClient(tx *generated.Tx) *generated.OrderItemProcessClient {
	if tx != nil {
		return tx.OrderItemProcess
	}
	return r.db.OrderItemProcess
}
