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
	db                    *generated.Client
	deps                  *module.ModuleDeps[config.ModuleConfig]
	cfMgr                 *customfields.Manager
	orderItemProcessRepo  OrderItemProcessRepository
	orderItemProductRepo  OrderItemProductRepository
	orderItemMaterialRepo OrderItemMaterialRepository
}

func NewOrderItemRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) OrderItemRepository {
	orderItemProcessRepo := NewOrderItemProcessRepository(db, deps, cfMgr)
	orderItemProductRepo := NewOrderItemProductRepository(db)
	orderItemMaterialRepo := NewOrderItemMaterialRepository(db)

	return &orderItemRepository{
		db:                    db,
		deps:                  deps,
		cfMgr:                 cfMgr,
		orderItemProcessRepo:  orderItemProcessRepo,
		orderItemProductRepo:  orderItemProductRepo,
		orderItemMaterialRepo: orderItemMaterialRepo,
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

	// Products
	if err := r.orderItemProductRepo.Load(ctx, dto); err != nil {
		return nil, err
	}

	// Materials
	if err := r.orderItemMaterialRepo.LoadConsumable(ctx, dto); err != nil {
		return nil, err
	}
	if err := r.orderItemMaterialRepo.LoadLoaner(ctx, dto); err != nil {
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
	productTotal, err := r.orderItemProductRepo.GetTotalPriceByOrderItemID(ctx, orderItemID)
	if err != nil {
		return 0, nil
	}

	consumableMaterialTotal, err := r.orderItemMaterialRepo.GetConsumableTotalPriceByOrderItemID(ctx, orderItemID)
	if err != nil {
		return 0, err
	}

	total := productTotal + consumableMaterialTotal
	return total, nil
}

func (r *orderItemRepository) GetTotalPriceByOrderID(ctx context.Context, orderID int64) (float64, error) {
	productTotal, err := r.orderItemProductRepo.GetTotalPriceByOrderID(ctx, orderID)
	if err != nil {
		return 0, nil
	}
	consumableMaterialTotal, err := r.orderItemMaterialRepo.GetConsumableTotalPriceByOrderID(ctx, orderID)
	if err != nil {
		return 0, nil
	}
	total := productTotal + consumableMaterialTotal
	return total, nil
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

func (r *orderItemRepository) applyTotalPrice(dto *model.OrderItemDTO, totalPrices ...*float64) {
	if dto == nil {
		return
	}

	if len(totalPrices) == 0 {
		dto.TotalPrice = nil
		return
	}

	var sum float64
	for _, p := range totalPrices {
		if p == nil {
			continue
		}
		sum += *p
	}

	dto.TotalPrice = &sum
}

// -- general functions
func (r *orderItemRepository) Create(ctx context.Context, tx *generated.Tx, input *model.OrderItemUpsertDTO) (*model.OrderItemDTO, error) {
	in := &input.DTO

	products := r.orderItemProductRepo.CollectProducts(in)
	totalPriceProduct := r.orderItemProductRepo.CalculateTotalPrice(products)
	consumableMaterials := r.orderItemMaterialRepo.CollectConsumableMaterials(in)
	totalPriceConsumableMaterial := r.orderItemMaterialRepo.CalculateConsumableTotalPrice(consumableMaterials)
	r.applyTotalPrice(in, totalPriceProduct, totalPriceConsumableMaterial)

	// order item - ParentItemID + RemakeCount
	prev, errLatest := r.GetLatestByOrderID(ctx, in.OrderID)
	if errLatest == nil && prev != nil {
		in.ParentItemID = &prev.ID
		in.RemakeCount = prev.RemakeCount + 1
	} else {
		in.ParentItemID = nil
		in.RemakeCount = 0
	}

	// order item - code
	if in.Code == nil || *in.Code == "" {
		if in.RemakeCount > 0 {
			seq, errSeq := r.getNextItemSeq(ctx, in.OrderID)
			if errSeq != nil {
				return nil, errSeq
			}

			alpha := utils.AlphabetSeq(seq)

			code := fmt.Sprintf("%s%s", alpha, *in.CodeOriginal)
			in.Code = &code
		} else {
			in.Code = in.CodeOriginal
		}
	}

	q := tx.OrderItem.Create().
		SetCode(*in.Code)
	if in.ParentItemID != nil {
		q.SetParentItemID(*in.ParentItemID)
	}
	q.SetRemakeCount(in.RemakeCount).
		SetOrderID(in.OrderID).
		SetNillableCodeOriginal(in.CodeOriginal).
		SetNillableTotalPrice(in.TotalPrice)

	// metadata

	// new order is as `received`
	cf := maps.Clone(in.CustomFields)
	cf["status"] = "received"
	in.CustomFields = cf

	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err := customfields.PrepareCustomFields(ctx,
			r.cfMgr,
			*input.Collections,
			in.CustomFields,
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

	out := mapper.MapAs[*generated.OrderItem, *model.OrderItemDTO](entity)

	// Products
	createdProducts, err := r.orderItemProductRepo.Sync(ctx, tx, entity.OrderID, entity.ID, products)
	if err != nil {
		return nil, err
	}
	out.Products = createdProducts

	// Consumable Materials
	createdConsumableMaterials, err := r.orderItemMaterialRepo.SyncConsumable(ctx, tx, entity.OrderID, entity.ID, consumableMaterials)
	if err != nil {
		return nil, err
	}
	out.ConsumableMaterials = createdConsumableMaterials

	// Loaner Materials
	loanerMaterials := r.orderItemMaterialRepo.CollectLoanerMaterials(in)
	createdLoanerMaterials, err := r.orderItemMaterialRepo.SyncLoaner(ctx, tx, entity.OrderID, entity.ID, loanerMaterials)
	if err != nil {
		return nil, err
	}
	out.LoanerMaterials = createdLoanerMaterials

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

	return out, nil
}

func (r *orderItemRepository) Update(ctx context.Context, tx *generated.Tx, input *model.OrderItemUpsertDTO) (*model.OrderItemDTO, error) {
	dto := &input.DTO
	products := r.orderItemProductRepo.CollectProducts(dto)
	totalPriceProduct := r.orderItemProductRepo.CalculateTotalPrice(products)
	consumableMaterials := r.orderItemMaterialRepo.CollectConsumableMaterials(dto)
	totalPriceConsumableMaterial := r.orderItemMaterialRepo.CalculateConsumableTotalPrice(consumableMaterials)
	r.applyTotalPrice(dto, totalPriceProduct, totalPriceConsumableMaterial)

	var primaryProductID int
	if len(products) > 0 {
		primaryProductID = products[0].ProductID
	}

	q := tx.OrderItem.UpdateOneID(dto.ID).
		SetNillableCode(dto.Code).
		SetNillableTotalPrice(dto.TotalPrice)

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

	out := mapper.MapAs[*generated.OrderItem, *model.OrderItemDTO](entity)

	// Products
	createdProducts, err := r.orderItemProductRepo.Sync(ctx, tx, entity.OrderID, entity.ID, products)
	if err != nil {
		return nil, err
	}
	out.Products = createdProducts

	// Consumable Materials
	createdConsumableMaterials, err := r.orderItemMaterialRepo.SyncConsumable(ctx, tx, entity.OrderID, entity.ID, consumableMaterials)
	if err != nil {
		return nil, err
	}
	out.ConsumableMaterials = createdConsumableMaterials

	// Loaner Materials
	loanerMaterials := r.orderItemMaterialRepo.CollectLoanerMaterials(dto)
	createdLoanerMaterials, err := r.orderItemMaterialRepo.SyncLoaner(ctx, tx, entity.OrderID, entity.ID, loanerMaterials)
	if err != nil {
		return nil, err
	}
	out.LoanerMaterials = createdLoanerMaterials

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
		out.OrderItemProcesses = oipOut
	}

	return out, nil
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

	// Products
	if err := r.orderItemProductRepo.Load(ctx, dto); err != nil {
		return nil, err
	}

	// Consumable Materials
	if err := r.orderItemMaterialRepo.LoadConsumable(ctx, dto); err != nil {
		return nil, err
	}

	// Loaner Materials
	if err := r.orderItemMaterialRepo.LoadLoaner(ctx, dto); err != nil {
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
	// Consumable Materials
	if err := r.orderItemMaterialRepo.LoadConsumable(ctx, list.Items...); err != nil {
		return list, err
	}

	// Loaner Materials
	if err := r.orderItemMaterialRepo.LoadLoaner(ctx, list.Items...); err != nil {
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
	return res, nil
}

func (r *orderItemRepository) Delete(ctx context.Context, id int64) error {
	return r.db.OrderItem.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
