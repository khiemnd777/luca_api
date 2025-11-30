package repository

import (
	"context"
	"fmt"
	"maps"
	"sort"

	"entgo.io/ent/dialect/sql"
	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemprocess"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/process"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/productprocess"
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
		productID int,
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

func (r *orderItemProcessRepository) CreateManyByProductID(
	ctx context.Context,
	tx *generated.Tx,
	orderItemID int64,
	orderID int64,
	orderCode *string,
	productID int,
) ([]*model.OrderItemProcessDTO, error) {
	processes, err := r.GetRawProcessesByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}

	if len(processes) == 0 {
		return []*model.OrderItemProcessDTO{}, nil
	}

	inputs := make([]*model.OrderItemProcessUpsertDTO, 0, len(processes))

	for i, p := range processes {
		// StepNumber = index + 1
		step := i + 1

		// Process name
		var pname *string
		if p.Name != nil {
			pname = p.Name
		}

		cf := map[string]any{}
		for k, v := range p.CustomFields {
			cf[k] = v
		}
		if _, ok := cf["status"]; !ok {
			cf["status"] = "waiting"
		}

		col := []string{"order-item-process"}

		dto := &model.OrderItemProcessUpsertDTO{
			DTO: model.OrderItemProcessDTO{
				OrderID:      &orderID,
				OrderItemID:  orderItemID,
				OrderCode:    orderCode,
				ProcessName:  pname,
				StepNumber:   step,
				CustomFields: cf,
			},
			Collections: &col,
		}

		inputs = append(inputs, dto)
	}

	// 3) Tạo hàng loạt
	return r.CreateMany(ctx, tx, inputs)
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

	q := tx.OrderItemProcess.Create().
		SetOrderItemID(dto.OrderItemID).
		SetNillableProcessName(dto.ProcessName).
		SetStepNumber(dto.StepNumber).
		SetNillableAssignedID(dto.AssignedID).
		SetNillableAssignedName(dto.AssignedName)

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

	err = relation.Upsert1(ctx, tx, "orderitemprocess", entity, &input.DTO, dto)
	if err != nil {
		return nil, err
	}

	_, err = relation.UpsertM2M(ctx, tx, "orderitemprocess", entity, input.DTO, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
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
		SetNillableAssignedID(dto.AssignedID).
		SetNillableAssignedName(dto.AssignedName).
		SetNillableNote(dto.Note).
		SetNillableStartedAt(dto.StartedAt).
		SetNillableCompletedAt(dto.CompletedAt)

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

	err = relation.Upsert1(ctx, tx, "orderitemprocess", entity, &input.DTO, dto)
	if err != nil {
		return nil, err
	}

	_, err = relation.UpsertM2M(ctx, tx, "orderitemprocess", entity, input.DTO, dto)
	if err != nil {
		return nil, err
	}

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
	processes, err := db.Process.
		Query().
		Where(
			process.HasProductsWith(
				productprocess.ProductIDEQ(productID),
			),
			process.DeletedAtIsNil(),
		).
		All(ctx)

	if err != nil {
		return nil, err
	}

	dtos := mapper.MapListAs[*generated.Process, *model.ProcessDTO](processes)
	return dtos, nil
}
