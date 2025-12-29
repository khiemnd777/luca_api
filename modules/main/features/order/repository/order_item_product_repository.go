package repository

import (
	"context"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitem"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemproduct"
	"github.com/khiemnd777/andy_api/shared/mapper"
)

type OrderItemProductRepository interface {
	CollectProducts(dto *model.OrderItemDTO) []*model.OrderItemProductDTO
	CalculateTotalPrice(products []*model.OrderItemProductDTO) *float64
	Sync(
		ctx context.Context,
		tx *generated.Tx,
		orderID,
		orderItemID int64,
		products []*model.OrderItemProductDTO,
	) ([]*model.OrderItemProductDTO, error)
	Load(ctx context.Context, items ...*model.OrderItemDTO) error
	GetTotalPriceByOrderItemID(ctx context.Context, orderItemID int64) (float64, error)
	GetTotalPriceByOrderID(ctx context.Context, tx *generated.Tx, orderID int64) (float64, error)
}

type orderItemProductRepository struct {
	db *generated.Client
}

func NewOrderItemProductRepository(db *generated.Client) OrderItemProductRepository {
	return &orderItemProductRepository{db: db}
}

func (r *orderItemProductRepository) CollectProducts(dto *model.OrderItemDTO) []*model.OrderItemProductDTO {
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

		qty := r.normalizeQuantity(product.Quantity)
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

func (r *orderItemProductRepository) CalculateTotalPrice(products []*model.OrderItemProductDTO) *float64 {
	var total float64
	hasPrice := false

	for _, product := range products {
		if product == nil || product.RetailPrice == nil {
			continue
		}
		qty := r.normalizeQuantity(product.Quantity)
		total += *product.RetailPrice * float64(qty)
		hasPrice = true
	}

	if !hasPrice {
		return nil
	}

	return &total
}

func (r *orderItemProductRepository) Sync(
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

		qty := r.normalizeQuantity(product.Quantity)
		create := tx.OrderItemProduct.Create().
			SetOrderID(orderID).
			SetOrderItemID(orderItemID).
			SetProductID(product.ProductID).
			SetQuantity(qty).
			SetNillableProductCode(product.ProductCode).
			SetNillableRetailPrice(product.RetailPrice)

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

func (r *orderItemProductRepository) Load(ctx context.Context, items ...*model.OrderItemDTO) error {
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

func (r *orderItemProductRepository) GetTotalPriceByOrderItemID(ctx context.Context, orderItemID int64) (float64, error) {
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
		qty := r.normalizeQuantity(product.Quantity)
		total += *product.RetailPrice * float64(qty)
	}

	return total, nil
}

func (r *orderItemProductRepository) GetTotalPriceByOrderID(ctx context.Context, tx *generated.Tx, orderID int64) (float64, error) {
	var oipC *generated.OrderItemProductClient
	if tx != nil {
		oipC = tx.OrderItemProduct
	} else {
		oipC = r.db.OrderItemProduct
	}
	products, err := oipC.
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
		qty := r.normalizeQuantity(product.Quantity)
		total += *product.RetailPrice * float64(qty)
	}

	return total, nil
}

func (r *orderItemProductRepository) normalizeQuantity(quantity int) int {
	if quantity <= 0 {
		return 1
	}
	return quantity
}
