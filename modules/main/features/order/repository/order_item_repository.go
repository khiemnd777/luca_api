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
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemproduct"
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
}

func NewOrderItemRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) OrderItemRepository {
	orderItemProcessRepo := NewOrderItemProcessRepository(db, deps, cfMgr)
	return &orderItemRepository{
		db:                   db,
		deps:                 deps,
		cfMgr:                cfMgr,
		orderItemProcessRepo: orderItemProcessRepo,
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
	if err := r.loadProducts(ctx, dto); err != nil {
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
	products, err := r.db.OrderItemProduct.
		Query().
		Where(orderitemproduct.OrderItemIDEQ(orderItemID)).
		Select(orderitemproduct.FieldQuantity, orderitemproduct.FieldRetailPrice).
		All(ctx)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, product := range products {
		if product == nil || product.RetailPrice == nil {
			continue
		}
		qty := normalizeQuantity(product.Quantity)
		total += *product.RetailPrice * float64(qty)
	}

	return total, nil
}

func (r *orderItemRepository) GetTotalPriceByOrderID(ctx context.Context, orderID int64) (float64, error) {
	products, err := r.db.OrderItemProduct.
		Query().
		Where(
			orderitemproduct.OrderIDEQ(orderID),
			orderitemproduct.HasOrderItemWith(orderitem.DeletedAtIsNil()),
		).
		Select(orderitemproduct.FieldQuantity, orderitemproduct.FieldRetailPrice).
		All(ctx)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, product := range products {
		if product == nil || product.RetailPrice == nil {
			continue
		}
		qty := normalizeQuantity(product.Quantity)
		total += *product.RetailPrice * float64(qty)
	}

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

func (r *orderItemRepository) collectProducts(dto *model.OrderItemDTO) []*model.OrderItemProductDTO {
	if dto == nil || len(dto.Products) == 0 {
		return nil
	}

	out := make([]*model.OrderItemProductDTO, 0, len(dto.Products))
	seen := make(map[int]struct{}, len(dto.Products))

	for _, product := range dto.Products {
		if product == nil || product.ProductID == 0 {
			continue
		}
		if _, ok := seen[product.ProductID]; ok {
			continue
		}
		seen[product.ProductID] = struct{}{}

		qty := normalizeQuantity(product.Quantity)
		out = append(out, &model.OrderItemProductDTO{
			ID:          product.ID,
			ProductID:   product.ProductID,
			OrderItemID: product.OrderItemID,
			OrderID:     product.OrderID,
			Quantity:    qty,
			RetailPrice: product.RetailPrice,
		})
	}

	return out
}

func normalizeQuantity(quantity int) int {
	if quantity <= 0 {
		return 1
	}
	return quantity
}

func calculateTotalPrice(products []*model.OrderItemProductDTO) *float64 {
	var total float64
	hasPrice := false

	for _, product := range products {
		if product == nil || product.RetailPrice == nil {
			continue
		}
		qty := normalizeQuantity(product.Quantity)
		total += *product.RetailPrice * float64(qty)
		hasPrice = true
	}

	if !hasPrice {
		return nil
	}

	return &total
}

func (r *orderItemRepository) applyTotalPrice(dto *model.OrderItemDTO, totalPrice *float64) {
	if dto == nil {
		return
	}

	dto.TotalPrice = totalPrice
}

func (r *orderItemRepository) syncOrderItemProducts(
	ctx context.Context,
	tx *generated.Tx,
	orderID,
	orderItemID int64,
	products []*model.OrderItemProductDTO,
) ([]*model.OrderItemProductDTO, error) {
	_, err := tx.OrderItemProduct.Delete().
		Where(orderitemproduct.OrderItemIDEQ(orderItemID)).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, nil
	}

	bulk := make([]*generated.OrderItemProductCreate, 0, len(products))
	for _, product := range products {
		if product == nil || product.ProductID == 0 {
			continue
		}

		qty := normalizeQuantity(product.Quantity)
		create := tx.OrderItemProduct.Create().
			SetOrderID(orderID).
			SetOrderItemID(orderItemID).
			SetProductID(product.ProductID).
			SetQuantity(qty)

		if product.RetailPrice != nil {
			create.SetRetailPrice(*product.RetailPrice)
		}

		bulk = append(bulk, create)
	}

	if len(bulk) == 0 {
		return nil, nil
	}

	created, err := tx.OrderItemProduct.CreateBulk(bulk...).Save(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*model.OrderItemProductDTO, 0, len(created))
	for _, it := range created {
		out = append(out, mapper.MapAs[*generated.OrderItemProduct, *model.OrderItemProductDTO](it))
	}

	return out, nil
}

func (r *orderItemRepository) loadProducts(ctx context.Context, items ...*model.OrderItemDTO) error {
	if len(items) == 0 {
		return nil
	}

	itemIndex := make(map[int64]*model.OrderItemDTO, len(items))
	itemIDs := make([]int64, 0, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		itemIDs = append(itemIDs, it.ID)
		itemIndex[it.ID] = it
	}

	if len(itemIDs) == 0 {
		return nil
	}

	relations, err := r.db.OrderItemProduct.Query().
		Where(orderitemproduct.OrderItemIDIn(itemIDs...)).
		All(ctx)
	if err != nil {
		return err
	}

	for _, rel := range relations {
		if dto, ok := itemIndex[rel.OrderItemID]; ok {
			dto.Products = append(dto.Products, mapper.MapAs[*generated.OrderItemProduct, *model.OrderItemProductDTO](rel))
		}
	}

	return nil
}

// -- general functions
func (r *orderItemRepository) Create(ctx context.Context, tx *generated.Tx, input *model.OrderItemUpsertDTO) (*model.OrderItemDTO, error) {
	dto := &input.DTO
	products := r.collectProducts(dto)
	totalPrice := calculateTotalPrice(products)
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

	createdProducts, err := r.syncOrderItemProducts(ctx, tx, entity.OrderID, entity.ID, products)
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
	products := r.collectProducts(dto)
	totalPrice := calculateTotalPrice(products)
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

	createdProducts, err := r.syncOrderItemProducts(ctx, tx, entity.OrderID, entity.ID, products)
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
	if err := r.loadProducts(ctx, dto); err != nil {
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
	if err := r.loadProducts(ctx, list.Items...); err != nil {
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
	if err := r.loadProducts(ctx, res.Items...); err != nil {
		return res, err
	}
	return res, nil
}

func (r *orderItemRepository) Delete(ctx context.Context, id int64) error {
	return r.db.OrderItem.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
