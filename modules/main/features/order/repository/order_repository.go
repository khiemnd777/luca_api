package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/order"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type OrderRepository interface {
	ExistsByCode(ctx context.Context, code string) (bool, error)
	// -- general functions
	Create(ctx context.Context, input *model.OrderUpsertDTO) (*model.OrderDTO, error)
	Update(ctx context.Context, input *model.OrderUpsertDTO) (*model.OrderDTO, error)
	GetByID(ctx context.Context, id int64) (*model.OrderDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.OrderDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.OrderDTO], error)
	Delete(ctx context.Context, id int64) error
}

type orderRepository struct {
	db            *generated.Client
	deps          *module.ModuleDeps[config.ModuleConfig]
	cfMgr         *customfields.Manager
	orderItemRepo OrderItemRepository
}

func NewOrderRepository(
	db *generated.Client,
	deps *module.ModuleDeps[config.ModuleConfig],
	cfMgr *customfields.Manager,
) OrderRepository {
	return &orderRepository{
		db:            db,
		deps:          deps,
		cfMgr:         cfMgr,
		orderItemRepo: NewOrderItemRepository(db, deps, cfMgr),
	}
}

func (r *orderRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	return r.db.Order.
		Query().
		Where(
			order.CodeEQ(code),
			order.DeletedAtIsNil(),
		).
		Exist(ctx)
}

// -- helpers

func (r *orderRepository) createNewOrder(
	ctx context.Context,
	tx *generated.Tx,
	input *model.OrderUpsertDTO,
) (*model.OrderDTO, error) {

	dto := &input.DTO

	q := tx.Order.Create().
		SetNillableCode(dto.Code)

	// custom fields
	if input.Collections != nil && len(*input.Collections) > 0 {
		if _, err := customfields.PrepareCustomFields(
			ctx, r.cfMgr, *input.Collections, dto.CustomFields, q, false,
		); err != nil {
			return nil, err
		}
	}

	// save order
	orderEnt, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	// map back
	out := mapper.MapAs[*generated.Order, *model.OrderDTO](orderEnt)

	// create first-latest order item
	loi := input.DTO.LatestOrderItemUpsert
	loi.DTO.OrderID = out.ID
	loi.DTO.CodeOriginal = out.Code

	latest, err := r.orderItemRepo.Create(ctx, tx, loi)
	if err != nil {
		return nil, err
	}

	out.LatestOrderItem = latest

	// reassign latest order item -> order as cache to appear them on the table
	lstStatus := latest.CustomFields["status"].(string)
	lstPriority := latest.CustomFields["priority"].(string)

	_, err = orderEnt.
		Update().
		SetNillableCodeLatest(latest.Code).
		SetNillableStatusLatest(&lstStatus).
		SetNillablePriorityLatest(&lstPriority).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	// Assign latest ones to output
	out.CodeLatest = latest.Code
	out.StatusLatest = &lstStatus
	out.PriorityLatest = &lstPriority

	// relation
	if err := relation.Upsert1(ctx, tx, "order", orderEnt, &input.DTO, out); err != nil {
		return nil, err
	}
	if _, err := relation.UpsertM2M(ctx, tx, "order", orderEnt, input.DTO, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *orderRepository) upsertExistingOrder(
	ctx context.Context,
	tx *generated.Tx,
	input *model.OrderUpsertDTO,
) (*model.OrderDTO, error) {

	dto := &input.DTO

	// Load order theo code
	orderEnt, err := r.db.Order.
		Query().
		Where(order.CodeEQ(*dto.Code), order.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	// UPDATE ORDER (custom fields + m2m + 1)
	up := tx.Order.UpdateOneID(orderEnt.ID).
		SetNillableCode(dto.Code)

	if input.Collections != nil && len(*input.Collections) > 0 {
		if _, err := customfields.PrepareCustomFields(
			ctx, r.cfMgr, *input.Collections, dto.CustomFields, up, false,
		); err != nil {
			return nil, err
		}
	}

	orderEnt, err = up.Save(ctx)
	if err != nil {
		return nil, err
	}

	out := mapper.MapAs[*generated.Order, *model.OrderDTO](orderEnt)

	loi := input.DTO.LatestOrderItemUpsert
	loi.DTO.OrderID = out.ID
	loi.DTO.CodeOriginal = out.Code

	latest, err := r.orderItemRepo.Create(ctx, tx, loi)
	if err != nil {
		return nil, err
	}

	out.LatestOrderItem = latest

	// reassign latest order item -> order as cache to appear them on the table
	lstStatus := latest.CustomFields["status"].(string)
	lstPriority := latest.CustomFields["priority"].(string)

	_, err = orderEnt.
		Update().
		SetNillableCodeLatest(latest.Code).
		SetNillableStatusLatest(&lstStatus).
		SetNillablePriorityLatest(&lstPriority).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	// Assign latest ones to output
	out.CodeLatest = latest.Code
	out.StatusLatest = &lstStatus
	out.PriorityLatest = &lstPriority

	// relations
	if err := relation.Upsert1(ctx, tx, "order", orderEnt, &input.DTO, out); err != nil {
		return nil, err
	}
	if _, err := relation.UpsertM2M(ctx, tx, "order", orderEnt, input.DTO, out); err != nil {
		return nil, err
	}

	return out, nil
}

// -- general functions

func (r *orderRepository) Create(ctx context.Context, input *model.OrderUpsertDTO) (*model.OrderDTO, error) {
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

	dto := &input.DTO
	code := dto.Code

	exists, err := r.ExistsByCode(ctx, *code)
	if err != nil {
		return nil, err
	}

	if exists {
		up, err := r.upsertExistingOrder(ctx, tx, input)
		if err != nil {
			return nil, err
		}
		return up, nil
	}

	new, err := r.createNewOrder(ctx, tx, input)
	if err != nil {
		return nil, err
	}

	return new, nil
}

func (r *orderRepository) Update(ctx context.Context, input *model.OrderUpsertDTO) (*model.OrderDTO, error) {
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

	output := &input.DTO

	q := tx.Order.UpdateOneID(output.ID)

	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err = customfields.PrepareCustomFields(ctx,
			r.cfMgr,
			*input.Collections,
			output.CustomFields,
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

	output = mapper.MapAs[*generated.Order, *model.OrderDTO](entity)

	// latest order item
	latest, err := r.orderItemRepo.Update(ctx, tx, input.DTO.LatestOrderItemUpsert)
	if err != nil {
		return nil, err
	}

	// reassign latest order item -> order as cache to appear them on the table
	isLatest, err := r.orderItemRepo.IsLatestIfOrderID(ctx, entity.ID, latest.ID)
	if err != nil {
		return nil, err
	}
	if isLatest {
		lstStatus := latest.CustomFields["status"].(string)
		lstPriority := latest.CustomFields["priority"].(string)

		_, err = entity.
			Update().
			SetNillableCodeLatest(latest.Code).
			SetNillableStatusLatest(&lstStatus).
			SetNillablePriorityLatest(&lstPriority).
			Save(ctx)

		if err != nil {
			return nil, err
		}

		// Assign latest ones to output
		output.CodeLatest = latest.Code
		output.StatusLatest = &lstStatus
		output.PriorityLatest = &lstPriority
	}

	output.LatestOrderItem = latest

	// relation
	err = relation.Upsert1(ctx, tx, "order", entity, &input.DTO, output)
	if err != nil {
		return nil, err
	}

	_, err = relation.UpsertM2M(ctx, tx, "order", entity, input.DTO, output)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*model.OrderDTO, error) {
	q := r.db.Order.Query().
		Where(
			order.ID(id),
			order.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Order, *model.OrderDTO](entity)

	// latest order item
	latest, err := r.orderItemRepo.GetLatestByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("[GET] %v", latest))
	dto.LatestOrderItem = latest
	return dto, nil
}

func (r *orderRepository) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.OrderDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Order.Query().
			Where(order.DeletedAtIsNil()),
		query,
		order.Table,
		order.FieldID,
		order.FieldID,
		func(src []*generated.Order) []*model.OrderDTO {
			return mapper.MapListAs[*generated.Order, *model.OrderDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.OrderDTO]
		return zero, err
	}
	return list, nil
}

func (r *orderRepository) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.OrderDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Order.Query().
			Where(order.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(order.FieldCode),
		},
		query,
		order.Table,
		order.FieldID,
		order.FieldID,
		order.Or,
		func(src []*generated.Order) []*model.OrderDTO {
			return mapper.MapListAs[*generated.Order, *model.OrderDTO](src)
		},
	)
}

func (r *orderRepository) Delete(ctx context.Context, id int64) error {
	return r.db.Order.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
