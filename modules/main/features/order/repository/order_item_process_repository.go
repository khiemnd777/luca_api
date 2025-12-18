package repository

import (
	"context"
	"fmt"
	"maps"
	"sort"

	"entgo.io/ent/dialect/sql"
	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/categoryprocess"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemprocess"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/process"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/product"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
)

type OrderItemProcessRepository interface {
	CreateManyByProductID(
		ctx context.Context,
		tx *generated.Tx,
		orderItemID int64,
		orderID int64,
		orderCode *string,
		priority *string,
		productID int,
	) ([]*model.OrderItemProcessDTO, error)

	CreateManyByProductIDs(
		ctx context.Context,
		tx *generated.Tx,
		orderItemID int64,
		orderID int64,
		orderCode *string,
		priority *string,
		productIDs []int,
	) ([]*model.OrderItemProcessDTO, error)

	CreateMany(
		ctx context.Context,
		tx *generated.Tx,
		inputs []*model.OrderItemProcessUpsertDTO,
	) ([]*model.OrderItemProcessDTO, error)

	Create(
		ctx context.Context,
		tx *generated.Tx,
		input *model.OrderItemProcessUpsertDTO,
	) (*model.OrderItemProcessDTO, error)

	UpdateManyWithProps(
		ctx context.Context,
		tx *generated.Tx,
		id int64,
		propsFn func(prop *model.OrderItemProcessDTO) error,
	) ([]*model.OrderItemProcessDTO, error)

	UpdateMany(
		ctx context.Context,
		tx *generated.Tx,
		inputs []*model.OrderItemProcessUpsertDTO,
	) ([]*model.OrderItemProcessDTO, error)

	Update(
		ctx context.Context,
		tx *generated.Tx,
		id int64,
		input *model.OrderItemProcessUpsertDTO,
	) (*model.OrderItemProcessDTO, error)

	UpdateStatus(
		ctx context.Context,
		tx *generated.Tx,
		id int64,
		status string,
	) (*model.OrderItemProcessDTO, error)

	GetProcessesByOrderItemID(
		ctx context.Context,
		tx *generated.Tx,
		orderItemID int64,
	) ([]*model.OrderItemProcessDTO, error)

	GetProcessesByAssignedID(
		ctx context.Context,
		tx *generated.Tx,
		assignedID int64,
	) ([]*model.OrderItemProcessDTO, error)

	GetProcessesByOrderID(
		ctx context.Context,
		tx *generated.Tx,
		orderID int64,
	) ([]*model.OrderItemProcessDTO, error)

	GetRawProcessesByProductID(
		ctx context.Context,
		productID int,
	) ([]*model.ProcessDTO, error)
}

type orderItemProcessRepository struct {
	db    *generated.Client
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewOrderItemProcessRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) OrderItemProcessRepository {
	return &orderItemProcessRepository{db: db, deps: deps, cfMgr: cfMgr}
}

func (r *orderItemProcessRepository) CreateManyByProductIDs(
	ctx context.Context,
	tx *generated.Tx,
	orderItemID int64,
	orderID int64,
	orderCode *string,
	priority *string,
	productIDs []int,
) ([]*model.OrderItemProcessDTO, error) {
	if len(productIDs) == 0 {
		return []*model.OrderItemProcessDTO{}, nil
	}

	uniqueProcesses := make([]*model.ProcessDTO, 0)
	seenProcessIDs := make(map[int]struct{})
	seenProductIDs := make(map[int]struct{}, len(productIDs))

	for _, pid := range productIDs {
		if pid == 0 {
			continue
		}
		if _, ok := seenProductIDs[pid]; ok {
			continue
		}
		seenProductIDs[pid] = struct{}{}

		processes, err := r.GetRawProcessesByProductID(ctx, pid)
		if err != nil {
			return nil, err
		}

		for _, p := range processes {
			if p == nil {
				continue
			}
			if _, ok := seenProcessIDs[p.ID]; ok {
				continue
			}
			seenProcessIDs[p.ID] = struct{}{}
			uniqueProcesses = append(uniqueProcesses, p)
		}
	}

	if len(uniqueProcesses) == 0 {
		// Why uniqueProcesses == 0?
		return []*model.OrderItemProcessDTO{}, nil
	}

	inputs := make([]*model.OrderItemProcessUpsertDTO, 0, len(uniqueProcesses))
	col := []string{"order-item-process"}

	for i, p := range uniqueProcesses {
		cf := maps.Clone(p.CustomFields)
		if cf == nil {
			cf = make(map[string]any)
		}

		if _, ok := cf["status"]; !ok {
			cf["status"] = "waiting"
		}
		if _, ok := cf["priority"]; !ok && priority != nil {
			cf["priority"] = *priority
		}

		var pname *string
		if p.Name != nil {
			pname = p.Name
		}

		inputs = append(inputs, &model.OrderItemProcessUpsertDTO{
			DTO: model.OrderItemProcessDTO{
				OrderID:      &orderID,
				OrderItemID:  orderItemID,
				OrderCode:    orderCode,
				Color:        p.Color,
				SectionName:  p.SectionName,
				ProcessName:  pname,
				StepNumber:   i + 1,
				CustomFields: cf,
			},
			Collections: &col,
		})
	}

	out, err := r.CreateMany(ctx, tx, inputs)
	if err != nil {
		logger.Debug(
			"create order item processes: bulk create failed",
			"orderItemID", orderItemID,
			"orderID", orderID,
			"processCount", len(inputs),
			"error", err,
		)
		return nil, err
	}

	logger.Debug(
		"create order item processes: finished",
		"productIDs", productIDs,
		"seenProductIDs", seenProductIDs,
		"seenProcessIDs", seenProcessIDs,
		"uniqueProcesses", uniqueProcesses,
		"out", out,
	)

	return out, nil
}

func (r *orderItemProcessRepository) CreateManyByProductID(
	ctx context.Context,
	tx *generated.Tx,
	orderItemID int64,
	orderID int64,
	orderCode *string,
	priority *string,
	productID int,
) ([]*model.OrderItemProcessDTO, error) {
	return r.CreateManyByProductIDs(ctx, tx, orderItemID, orderID, orderCode, priority, []int{productID})
}

func (r *orderItemProcessRepository) CreateMany(
	ctx context.Context,
	tx *generated.Tx,
	inputs []*model.OrderItemProcessUpsertDTO,
) ([]*model.OrderItemProcessDTO, error) {

	if len(inputs) == 0 {
		return []*model.OrderItemProcessDTO{}, nil
	}

	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].DTO.StepNumber < inputs[j].DTO.StepNumber
	})

	out := make([]*model.OrderItemProcessDTO, 0, len(inputs))

	for _, in := range inputs {
		dto, err := r.Create(ctx, tx, in)
		if err != nil {
			return nil, err
		}
		out = append(out, dto)
	}

	return out, nil
}

func (r *orderItemProcessRepository) Create(ctx context.Context, tx *generated.Tx, input *model.OrderItemProcessUpsertDTO) (*model.OrderItemProcessDTO, error) {
	dto := &input.DTO

	// customfields
	if dto.CustomFields == nil {
		dto.CustomFields = make(map[string]any)
	}

	if _, exists := dto.CustomFields["status"]; !exists {
		dto.CustomFields["status"] = "waiting"
	}

	q := tx.OrderItemProcess.
		Create().
		SetOrderItemID(dto.OrderItemID).
		SetNillableOrderID(dto.OrderID).
		SetNillableProcessName(dto.ProcessName).
		SetStepNumber(dto.StepNumber).
		SetNillableAssignedID(dto.AssignedID).
		SetNillableAssignedName(dto.AssignedName).
		SetNillableColor(dto.Color).
		SetNillableSectionName(dto.SectionName)

	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err := customfields.PrepareCustomFields(ctx,
			r.cfMgr,
			*input.Collections,
			dto.CustomFields,
			q,
			false,
		)
		if err != nil {
			return nil, err
		}
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto = mapper.MapAs[*generated.OrderItemProcess, *model.OrderItemProcessDTO](entity)

	return dto, nil
}

func (r *orderItemProcessRepository) UpdateManyWithProps(
	ctx context.Context,
	tx *generated.Tx,
	id int64,
	propsFn func(prop *model.OrderItemProcessDTO) error,
) ([]*model.OrderItemProcessDTO, error) {
	poiList, err := r.GetProcessesByOrderItemID(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	for _, poi := range poiList {
		if propsFn != nil {
			err := propsFn(poi)
			if err != nil {
				return nil, err
			}
		}
	}
	col := []string{"order-item-process"}
	oipDTOs := make([]*model.OrderItemProcessUpsertDTO, 0, len(poiList))
	for _, poi := range poiList {
		oipDTOs = append(oipDTOs, &model.OrderItemProcessUpsertDTO{
			DTO:         *poi,
			Collections: &col,
		})
	}
	out, err := r.UpdateMany(ctx, tx, oipDTOs)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *orderItemProcessRepository) UpdateMany(
	ctx context.Context,
	tx *generated.Tx,
	inputs []*model.OrderItemProcessUpsertDTO,
) ([]*model.OrderItemProcessDTO, error) {

	out := make([]*model.OrderItemProcessDTO, 0, len(inputs))

	for _, in := range inputs {
		id := in.DTO.ID
		if id == 0 {
			return nil, fmt.Errorf("missing ID for update")
		}

		dto, err := r.Update(ctx, tx, id, in)
		if err != nil {
			return nil, err
		}

		out = append(out, dto)
	}

	return out, nil
}

func (r *orderItemProcessRepository) Update(
	ctx context.Context,
	tx *generated.Tx,
	id int64,
	input *model.OrderItemProcessUpsertDTO,
) (*model.OrderItemProcessDTO, error) {

	dto := &input.DTO

	existing, err := tx.OrderItemProcess.
		Query().
		Where(orderitemprocess.IDEQ(id)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	q := tx.OrderItemProcess.
		UpdateOne(existing).
		SetNillableOrderID(dto.OrderID).
		SetNillableAssignedID(dto.AssignedID).
		SetNillableAssignedName(dto.AssignedName).
		SetNillableNote(dto.Note).
		SetNillableStartedAt(dto.StartedAt).
		SetNillableCompletedAt(dto.CompletedAt).
		SetNillableColor(dto.Color).
		SetNillableSectionName(dto.SectionName)

	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err := customfields.PrepareCustomFields(
			ctx,
			r.cfMgr,
			*input.Collections,
			dto.CustomFields,
			q,
			false,
		)
		if err != nil {
			return nil, err
		}
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	out := mapper.MapAs[*generated.OrderItemProcess, *model.OrderItemProcessDTO](entity)

	return out, nil
}

func (r *orderItemProcessRepository) UpdateStatus(
	ctx context.Context,
	tx *generated.Tx,
	id int64,
	status string,
) (*model.OrderItemProcessDTO, error) {

	oip, err := tx.OrderItemProcess.
		Query().
		Where(orderitemprocess.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	cf := maps.Clone(oip.CustomFields)
	cf["status"] = status

	entity, err := tx.OrderItemProcess.
		UpdateOneID(id).
		SetCustomFields(cf).
		Save(ctx)

	if err != nil {
		return nil, err
	}
	out := mapper.MapAs[*generated.OrderItemProcess, *model.OrderItemProcessDTO](entity)

	return out, nil
}

func (r *orderItemProcessRepository) GetProcessesByOrderItemID(
	ctx context.Context,
	tx *generated.Tx,
	orderItemID int64,
) ([]*model.OrderItemProcessDTO, error) {
	var oipC *generated.OrderItemProcessClient
	if tx != nil {
		oipC = tx.OrderItemProcess
	} else {
		oipC = r.db.OrderItemProcess
	}
	items, err := oipC.
		Query().
		Where(
			orderitemprocess.OrderItemID(orderItemID),
		).
		Order(
			orderitemprocess.ByStepNumber(
				sql.OrderAsc(),
			),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := mapper.MapListAs[*generated.OrderItemProcess, *model.OrderItemProcessDTO](items)
	return out, nil
}

func (r *orderItemProcessRepository) GetProcessesByAssignedID(
	ctx context.Context,
	tx *generated.Tx,
	assignedID int64,
) ([]*model.OrderItemProcessDTO, error) {
	var oipC *generated.OrderItemProcessClient
	if tx != nil {
		oipC = tx.OrderItemProcess
	} else {
		oipC = r.db.OrderItemProcess
	}
	items, err := oipC.
		Query().
		Where(
			orderitemprocess.AssignedID(assignedID),
		).
		Order(
			orderitemprocess.ByStepNumber(
				sql.OrderAsc(),
			),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := mapper.MapListAs[*generated.OrderItemProcess, *model.OrderItemProcessDTO](items)
	return out, nil
}

func (r *orderItemProcessRepository) GetProcessesByOrderID(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
) ([]*model.OrderItemProcessDTO, error) {

	items, err := tx.OrderItemProcess.
		Query().
		Where(
			orderitemprocess.OrderID(orderID),
		).
		Order(
			orderitemprocess.ByStepNumber(
				sql.OrderAsc(),
			),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := mapper.MapListAs[*generated.OrderItemProcess, *model.OrderItemProcessDTO](items)
	return out, nil
}

func (r *orderItemProcessRepository) GetRawProcessesByProductID(
	ctx context.Context,
	productID int,
) ([]*model.ProcessDTO, error) {
	db := r.db

	prd, err := db.Product.
		Query().
		Where(
			product.IDEQ(productID),
			product.DeletedAtIsNil(),
		).
		Select(product.FieldCategoryID).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	if prd.CategoryID == nil {
		return []*model.ProcessDTO{}, nil
	}

	processes, err := db.Process.
		Query().
		Where(
			process.HasCategoriesWith(
				categoryprocess.CategoryIDEQ(*prd.CategoryID),
			),
			process.DeletedAtIsNil(),
		).
		Order(func(s *sql.Selector) {
			cp := sql.Table(categoryprocess.Table)

			s.Join(cp).
				On(
					s.C(process.FieldID),
					cp.C(categoryprocess.FieldProcessID),
				)

			s.OrderBy(cp.C(categoryprocess.FieldDisplayOrder))
		}).
		All(ctx)

	if err != nil {
		return nil, err
	}

	dtos := mapper.MapListAs[*generated.Process, *model.ProcessDTO](processes)
	return dtos, nil
}
