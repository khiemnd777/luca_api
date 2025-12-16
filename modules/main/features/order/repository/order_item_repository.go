package repository

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
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
	GetHistoricalByOrderIDAndOrderItemID(ctx context.Context, orderID, orderItemID int64) ([]*model.OrderItemHistoricalDTO, error)
	GetTotalPriceByOrderItemID(ctx context.Context, orderItemID int64) (float64, error)
	GetTotalPriceByOrderID(ctx context.Context, orderID int64) (float64, error)
	// -- general functions
	Create(ctx context.Context, tx *generated.Tx, input *model.OrderItemUpsertDTO) (*model.OrderItemDTO, error)
	Update(ctx context.Context, tx *generated.Tx, input *model.OrderItemUpsertDTO) (*model.OrderItemDTO, error)
	GetByID(ctx context.Context, id int64) (*model.OrderItemDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.OrderItemDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.OrderItemDTO], error)
	Delete(ctx context.Context, id int64) error
}

type orderItemRepository struct {
	db                   *generated.Client
	deps                 *module.ModuleDeps[config.ModuleConfig]
	cfMgr                *customfields.Manager
	orderItemProcessRepo OrderItemProcessRepository
	orderItemProductRepo OrderItemProductRepository
}

func NewOrderItemRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) OrderItemRepository {
	orderItemProcessRepo := NewOrderItemProcessRepository(db, deps, cfMgr)
	orderItemProductRepo := NewOrderItemProductRepository(db)
	return &orderItemRepository{
		db:                   db,
		deps:                 deps,
		cfMgr:                cfMgr,
		orderItemProcessRepo: orderItemProcessRepo,
		orderItemProductRepo: orderItemProductRepo,
	}
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
	if err := r.orderItemProductRepo.Load(ctx, dto); err != nil {
		return nil, err
	}
	return dto, nil
}

func (r *orderItemRepository) GetHistoricalByOrderIDAndOrderItemID(
	ctx context.Context,
	orderID, orderItemID int64,
) ([]*model.OrderItemHistoricalDTO, error) {

	items, err := r.db.OrderItem.
		Query().
		Where(
			orderitem.OrderID(orderID),
			orderitem.DeletedAtIsNil(),
		).
		Order(generated.Desc(orderitem.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*model.OrderItemHistoricalDTO, 0, len(items))

	var latestID int64 = 0
	if len(items) > 0 {
		latestID = items[0].ID
	}

	for _, it := range items {
		id := it.ID

		isCurrent := (id == latestID)

		var isHighlight bool
		if orderItemID == 0 {
			isHighlight = (id == latestID)
		} else {
			isHighlight = (id == orderItemID)
		}

		out = append(out, &model.OrderItemHistoricalDTO{
			ID:          id,
			Code:        *it.Code,
			CreatedAt:   it.CreatedAt,
			IsCurrent:   isCurrent,
			IsHighlight: isHighlight,
		})
	}

	return out, nil
}

func (r *orderItemRepository) GetTotalPriceByOrderItemID(ctx context.Context, orderItemID int64) (float64, error) {
	return r.orderItemProductRepo.GetTotalPriceByOrderItemID(ctx, orderItemID)
}

func (r *orderItemRepository) GetTotalPriceByOrderID(ctx context.Context, orderID int64) (float64, error) {
	return r.orderItemProductRepo.GetTotalPriceByOrderID(ctx, orderID)
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

func (r *orderItemRepository) applyTotalPrice(dto *model.OrderItemDTO, totalPrice *float64) {
	if dto == nil {
		return
	}

	dto.TotalPrice = totalPrice
}

// -- general functions
func (r *orderItemRepository) Create(ctx context.Context, tx *generated.Tx, input *model.OrderItemUpsertDTO) (*model.OrderItemDTO, error) {
	dto := &input.DTO
	products := r.orderItemProductRepo.CollectProducts(dto)
	totalPrice := r.orderItemProductRepo.CalculateTotalPrice(products)
	r.applyTotalPrice(dto, totalPrice)

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
		if dto.RemakeCount > 0 {
			seq, errSeq := r.getNextItemSeq(ctx, dto.OrderID)
			if errSeq != nil {
				return nil, errSeq
			}

			alpha := utils.AlphabetSeq(seq)

			code := fmt.Sprintf("%s%s", alpha, *dto.CodeOriginal)
			dto.Code = &code
		} else {
			dto.Code = dto.CodeOriginal
		}
	}

	q := tx.OrderItem.Create().
		SetCode(*dto.Code)
	if dto.ParentItemID != nil {
		q.SetParentItemID(*dto.ParentItemID)
	}
	q.SetRemakeCount(dto.RemakeCount).
		SetOrderID(dto.OrderID).
		SetNillableCodeOriginal(dto.CodeOriginal).
		SetNillableTotalPrice(totalPrice)

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

	createdProducts, err := r.orderItemProductRepo.Sync(ctx, tx, entity.OrderID, entity.ID, products)
	if err != nil {
		return nil, err
	}

	dto = mapper.MapAs[*generated.OrderItem, *model.OrderItemDTO](entity)
	dto.Products = createdProducts

	// processes
	if len(products) > 0 {
		priority := utils.SafeGetString(entity.CustomFields, "priority")
		for _, product := range products {
			if product == nil || product.ProductID == 0 {
				continue
			}
			if _, err := r.orderItemProcessRepo.CreateManyByProductID(ctx, tx, entity.ID, entity.OrderID, entity.Code, &priority, product.ProductID); err != nil {
				return nil, err
			}
		}
	}

	return dto, nil
}

func (r *orderItemRepository) Update(ctx context.Context, tx *generated.Tx, input *model.OrderItemUpsertDTO) (*model.OrderItemDTO, error) {
	dto := &input.DTO
	products := r.orderItemProductRepo.CollectProducts(dto)
	totalPrice := r.orderItemProductRepo.CalculateTotalPrice(products)
	r.applyTotalPrice(dto, totalPrice)

	var primaryProductID int
	if len(products) > 0 {
		primaryProductID = products[0].ProductID
	}

	q := tx.OrderItem.UpdateOneID(dto.ID).
		SetNillableCode(dto.Code).
		SetNillableTotalPrice(totalPrice)

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

	createdProducts, err := r.orderItemProductRepo.Sync(ctx, tx, entity.OrderID, entity.ID, products)
	if err != nil {
		return nil, err
	}

	dto = mapper.MapAs[*generated.OrderItem, *model.OrderItemDTO](entity)
	dto.Products = createdProducts

	// processes
	if primaryProductID > 0 {
		priority := utils.SafeGetString(entity.CustomFields, "priority")
		oipOut, err := r.orderItemProcessRepo.UpdateManyWithProps(ctx, tx, entity.ID, func(prop *model.OrderItemProcessDTO) error {
			cf := maps.Clone(prop.CustomFields)
			if cf != nil {
				if priority != "" {
					cf["priority"] = priority
				}
				prop.CustomFields = cf
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		dto.OrderItemProcesses = oipOut
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
	if err := r.orderItemProductRepo.Load(ctx, dto); err != nil {
		return nil, err
	}

	// processes
	// prcs, err := r.orderItemProcessRepo.GetByOrderItemID(ctx, id)
	// if err != nil {
	// 	return nil, err
	// }
	// dto.OrderItemProcesses = prcs

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
	if err := r.orderItemProductRepo.Load(ctx, list.Items...); err != nil {
		return list, err
	}
	return list, nil
}

func (r *orderItemRepository) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.OrderItemDTO], error) {
	res, err := dbutils.Search(
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
	if err != nil {
		var zero dbutils.SearchResult[model.OrderItemDTO]
		return zero, err
	}
	if err := r.orderItemProductRepo.Load(ctx, res.Items...); err != nil {
		return res, err
	}
	return res, nil
}

func (r *orderItemRepository) Delete(ctx context.Context, id int64) error {
	return r.db.OrderItem.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
