package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitem"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type OrderItemRepository interface {
	IsLatest(ctx context.Context, orderItemID int64) (bool, error)
	IsLatestIfOrderID(ctx context.Context, orderID, orderItemID int64) (bool, error)
	GetLatestByOrderID(ctx context.Context, orderID int64) (*model.OrderItemDTO, error)
	// -- general functions
	Create(ctx context.Context, tx *generated.Tx, input *model.OrderItemUpsertDTO) (*model.OrderItemDTO, error)
	Update(ctx context.Context, tx *generated.Tx, input *model.OrderItemUpsertDTO) (*model.OrderItemDTO, error)
	GetByID(ctx context.Context, id int64) (*model.OrderItemDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.OrderItemDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.OrderItemDTO], error)
	Delete(ctx context.Context, id int64) error
}

type orderItemRepository struct {
	db    *generated.Client
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewOrderItemRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) OrderItemRepository {
	return &orderItemRepository{db: db, deps: deps, cfMgr: cfMgr}
}

func (r *orderItemRepository) IsLatest(ctx context.Context, orderItemID int64) (bool, error) {
	cur, err := r.db.OrderItem.
		Query().
		Where(
			orderitem.ID(orderItemID),
			orderitem.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return false, err
	}

	orderID := cur.OrderID

	latest, err := r.db.OrderItem.
		Query().
		Where(
			orderitem.OrderID(orderID),
			orderitem.DeletedAtIsNil(),
		).
		Order(generated.Desc(orderitem.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		return false, err
	}
	return latest.ID == orderItemID, nil
}

func (r *orderItemRepository) IsLatestIfOrderID(ctx context.Context, orderID, orderItemID int64) (bool, error) {
	latest, err := r.db.OrderItem.
		Query().
		Where(
			orderitem.OrderID(orderID),
			orderitem.DeletedAtIsNil(),
		).
		Order(generated.Desc(orderitem.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		return false, err
	}
	return latest.ID == orderItemID, nil
}

func (r *orderItemRepository) GetLatestByOrderID(ctx context.Context, orderID int64) (*model.OrderItemDTO, error) {
	itemEnt, err := r.db.OrderItem.
		Query().
		Where(
			orderitem.OrderID(orderID),
			orderitem.DeletedAtIsNil(),
		).
		Order(generated.Desc(orderitem.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.OrderItem, *model.OrderItemDTO](itemEnt)
	return dto, nil
}

// -- helpers
func (r *orderItemRepository) getNextItemSeq(ctx context.Context, orderID int64) (int, error) {
	count, err := r.db.OrderItem.
		Query().
		Where(
			orderitem.OrderID(orderID),
			orderitem.DeletedAtIsNil(),
		).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return count + 1, nil
}

// -- general functions
func (r *orderItemRepository) Create(ctx context.Context, tx *generated.Tx, input *model.OrderItemUpsertDTO) (*model.OrderItemDTO, error) {
	dto := &input.DTO

	// order item - ParentItemID + RemakeCount
	prev, errLatest := r.GetLatestByOrderID(ctx, dto.OrderID)
	if errLatest == nil && prev != nil {
		dto.ParentItemID = &prev.ID
		dto.RemakeCount = prev.RemakeCount + 1
	} else {
		dto.ParentItemID = nil
		dto.RemakeCount = 0
	}

	// order item - code
	if dto.Code == nil || *dto.Code == "" {
		seq, errSeq := r.getNextItemSeq(ctx, dto.OrderID)
		if errSeq != nil {
			return nil, errSeq
		}

		alpha := utils.AlphabetSeq(seq)

		code := fmt.Sprintf("%s-%s", alpha, *dto.CodeOriginal)
		dto.Code = &code
	}

	q := tx.OrderItem.Create().
		SetCode(*dto.Code)
	if dto.ParentItemID != nil {
		q.SetParentItemID(*dto.ParentItemID)
	}
	q.SetRemakeCount(dto.RemakeCount)
	q.SetOrderID(dto.OrderID)
	q.SetNillableCodeOriginal(dto.CodeOriginal)

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

	dto = mapper.MapAs[*generated.OrderItem, *model.OrderItemDTO](entity)

	err = relation.Upsert1(ctx, tx, "orderitem", entity, &input.DTO, dto)
	if err != nil {
		return nil, err
	}

	_, err = relation.UpsertM2M(ctx, tx, "orderitem", entity, input.DTO, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *orderItemRepository) Update(ctx context.Context, tx *generated.Tx, input *model.OrderItemUpsertDTO) (*model.OrderItemDTO, error) {
	dto := &input.DTO

	q := tx.OrderItem.UpdateOneID(dto.ID).
		SetNillableCode(dto.Code)

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

	dto = mapper.MapAs[*generated.OrderItem, *model.OrderItemDTO](entity)

	err = relation.Upsert1(ctx, tx, "orderitem", entity, &input.DTO, dto)
	if err != nil {
		return nil, err
	}

	_, err = relation.UpsertM2M(ctx, tx, "orderitem", entity, input.DTO, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *orderItemRepository) GetByID(ctx context.Context, id int64) (*model.OrderItemDTO, error) {
	q := r.db.OrderItem.Query().
		Where(
			orderitem.ID(id),
			orderitem.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.OrderItem, *model.OrderItemDTO](entity)
	return dto, nil
}

func (r *orderItemRepository) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.OrderItemDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.OrderItem.Query().
			Where(orderitem.DeletedAtIsNil()),
		query,
		orderitem.Table,
		orderitem.FieldID,
		orderitem.FieldID,
		func(src []*generated.OrderItem) []*model.OrderItemDTO {
			return mapper.MapListAs[*generated.OrderItem, *model.OrderItemDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.OrderItemDTO]
		return zero, err
	}
	return list, nil
}

func (r *orderItemRepository) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.OrderItemDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.OrderItem.Query().
			Where(orderitem.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(orderitem.FieldCode),
		},
		query,
		orderitem.Table,
		orderitem.FieldID,
		orderitem.FieldID,
		orderitem.Or,
		func(src []*generated.OrderItem) []*model.OrderItemDTO {
			return mapper.MapListAs[*generated.OrderItem, *model.OrderItemDTO](src)
		},
	)
}

func (r *orderItemRepository) Delete(ctx context.Context, id int64) error {
	return r.db.OrderItem.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
