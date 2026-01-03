package repository

import (
	"context"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitem"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemproduct"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/product"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/mapper"
)

type OrderItemProductRepository interface {
	PrepareProducts(dto *model.OrderItemDTO) []*model.OrderItemProductDTO
	CalculateTotalPrice(products []*model.OrderItemProductDTO) *float64

	Sync(
		ctx context.Context,
		tx *generated.Tx,
		orderID int64,
		orderItemID int64,
		products []*model.OrderItemProductDTO,
	) ([]*model.OrderItemProductDTO, error)

	GetProductsByOrderID(ctx context.Context, orderID int64) ([]*model.OrderItemProductDTO, error)
	Load(ctx context.Context, items ...*model.OrderItemDTO) error
	PrepareForRemake(
		ctx context.Context,
		items ...*model.OrderItemDTO,
	) error
	GetTotalPriceByOrderItemID(ctx context.Context, orderItemID int64) (float64, error)
	GetTotalPriceByOrderID(ctx context.Context, tx *generated.Tx, orderID int64) (float64, error)
}

type orderItemProductRepository struct {
	db *generated.Client
}

func NewOrderItemProductRepository(db *generated.Client) OrderItemProductRepository {
	return &orderItemProductRepository{db: db}
}

func (r *orderItemProductRepository) PrepareProducts(
	dto *model.OrderItemDTO,
) []*model.OrderItemProductDTO {
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
			ID:                  product.ID,
			ProductID:           product.ProductID,
			OrderItemID:         product.OrderItemID,
			OriginalOrderItemID: product.OriginalOrderItemID,
			IsCloneable:         product.IsCloneable,
			OrderID:             product.OrderID,
			Quantity:            qty,
			RetailPrice:         product.RetailPrice,
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

// Sync V2 rules:
// 1) ALWAYS write to current orderItemID first (orderItemID owns its own rows)
// 2) If IsCloneable == true  -> sync UP to parent (OriginalOrderItemID)
// 3) If IsCloneable == false -> sync DOWN to children (find derived by OriginalOrderItemID == orderItemID)
//
// Notes:
//   - IsCloneable nil is treated as false (sync down).
//   - When writing current rows:
//     OriginalOrderItemID = dto.OriginalOrderItemID if provided, else = orderItemID (self/source)
func (r *orderItemProductRepository) Sync(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	orderItemID int64,
	products []*model.OrderItemProductDTO,
) ([]*model.OrderItemProductDTO, error) {

	logger.Debug("SyncProductV2: start",
		"orderID", orderID,
		"orderItemID", orderItemID,
		"inputCount", len(products),
	)

	// -----------------------------
	// 1) Normalize + partition by IsCloneable policy
	// -----------------------------
	currentProducts := make([]*model.OrderItemProductDTO, 0, len(products))

	// clone-to-parent grouped by parent OrderItemID
	cloneToParent := make(map[int64][]*model.OrderItemProductDTO)

	// clone-to-children list (source for fan-out)
	cloneToChildren := make([]*model.OrderItemProductDTO, 0)

	for _, p := range products {
		if p == nil || p.ProductID == 0 {
			continue
		}

		// Always belongs to current orderItemID
		currentProducts = append(currentProducts, p)

		isCloneable := p.IsCloneable != nil && *p.IsCloneable

		if isCloneable {
			// Sync UP requires a parent
			if p.OriginalOrderItemID != nil && *p.OriginalOrderItemID != orderItemID {
				parentOID := *p.OriginalOrderItemID
				cloneToParent[parentOID] = append(cloneToParent[parentOID], p)
			}
		} else {
			// Sync DOWN from current to its children
			cloneToChildren = append(cloneToChildren, p)
		}
	}

	logger.Debug("SyncProductV2: partition input",
		"currentCount", len(currentProducts),
		"cloneToParentGroups", len(cloneToParent),
		"cloneToChildrenCount", len(cloneToChildren),
	)

	// -----------------------------
	// 2) ALWAYS write CURRENT orderItemID
	// -----------------------------
	logger.Debug("SyncProductV2: write current (replace)",
		"orderItemID", orderItemID,
		"count", len(currentProducts),
	)

	if err := r.replaceProductsForOrderItem(ctx, tx, orderID, orderItemID, currentProducts); err != nil {
		return nil, err
	}

	// -----------------------------
	// 3) IsCloneable=true -> Sync UP to parent (fan-in)
	// -----------------------------
	for parentOID, items := range cloneToParent {
		logger.Debug("SyncProductV2: clone UP to parent",
			"fromOrderItemID", orderItemID,
			"toParentOrderItemID", parentOID,
			"count", len(items),
		)

		// Reuse your existing semantics: derived overrides parent
		if err := r.syncFromDerived(
			ctx,
			tx,
			orderID,
			orderItemID, // derived/current
			parentOID,   // parent/source
			items,
		); err != nil {
			return nil, err
		}

		// SyncProductV2: parent -> all children
		logger.Debug("SyncProductV2: parent -> all children",
			"source", parentOID,
			"count", len(items),
		)

		if err := r.syncFromSource(
			ctx,
			tx,
			orderID,
			parentOID,
			items,
		); err != nil {
			return nil, err
		}
	}

	// -----------------------------
	// 4) IsCloneable=false -> Sync DOWN to children (fan-out)
	// -----------------------------
	if len(cloneToChildren) > 0 {
		logger.Debug("SyncProductV2: clone DOWN to children",
			"sourceOrderItemID", orderItemID,
			"count", len(cloneToChildren),
		)

		// Reuse your existing semantics: source cascades to children
		if err := r.syncFromSource(
			ctx,
			tx,
			orderID,
			orderItemID, // current as source for children
			cloneToChildren,
		); err != nil {
			return nil, err
		}
	}

	// -----------------------------
	// 5) Return CURRENT rows (must never be empty due to missing routing)
	// -----------------------------
	finalRows, err := tx.OrderItemProduct.
		Query().
		Where(orderitemproduct.OrderItemIDEQ(orderItemID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	logger.Debug("SyncProductV2: completed",
		"orderItemID", orderItemID,
		"finalCount", len(finalRows),
	)

	out := make([]*model.OrderItemProductDTO, 0, len(finalRows))
	for _, row := range finalRows {
		out = append(out, mapper.MapAs[*generated.OrderItemProduct, *model.OrderItemProductDTO](row))
	}

	return out, nil
}

// replaceProductsForOrderItem replaces all products of current orderItemID with the provided list.
// This guarantees the invariant: Sync(...) writes to current orderItemID.
func (r *orderItemProductRepository) replaceProductsForOrderItem(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	orderItemID int64,
	products []*model.OrderItemProductDTO,
) error {

	// Delete all current rows first
	if _, err := tx.OrderItemProduct.Delete().
		Where(orderitemproduct.OrderItemIDEQ(orderItemID)).
		Exec(ctx); err != nil {
		return err
	}

	if len(products) == 0 {
		return nil
	}

	bulk := make([]*generated.OrderItemProductCreate, 0, len(products))

	for _, p := range products {
		if p == nil || p.ProductID == 0 {
			continue
		}

		// OriginalOrderItemID semantics:
		// - If input provides a parent, keep it
		// - Otherwise mark as self/source
		origOID := orderItemID
		if p.OriginalOrderItemID != nil && *p.OriginalOrderItemID != 0 {
			origOID = *p.OriginalOrderItemID
		}

		c := tx.OrderItemProduct.Create().
			SetOrderID(orderID).
			SetOrderItemID(orderItemID).
			SetOriginalOrderItemID(origOID).
			SetProductID(p.ProductID).
			SetQuantity(p.Quantity).
			SetNillableRetailPrice(p.RetailPrice).
			SetNillableIsCloneable(p.IsCloneable)

		// Optional fields if your schema supports them (uncomment if applicable):
		// c.SetNillableProductCode(p.ProductCode)
		// c.SetNillableProductName(p.ProductName)
		// c.SetNillableOrderItemCode(p.OrderItemCode)

		bulk = append(bulk, c)
	}

	if len(bulk) == 0 {
		return nil
	}

	_, err := tx.OrderItemProduct.CreateBulk(bulk...).Save(ctx)
	return err
}

func (r *orderItemProductRepository) syncFromDerived(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	derivedOrderItemID int64,
	sourceOrderItemID int64,
	products []*model.OrderItemProductDTO,
) error {

	// IMPORTANT: only delete "owned/lineage" rows of the source
	// Keep inherited rows (OriginalOrderItemID != sourceOrderItemID)
	if _, err := tx.OrderItemProduct.Delete().
		Where(
			orderitemproduct.OrderItemIDEQ(sourceOrderItemID),
			orderitemproduct.OriginalOrderItemIDEQ(sourceOrderItemID),
		).
		Exec(ctx); err != nil {
		return err
	}

	if len(products) == 0 {
		return nil
	}

	var bulk []*generated.OrderItemProductCreate

	for _, p := range products {
		if p == nil || p.ProductID == 0 {
			continue
		}

		qty := r.normalizeQuantity(p.Quantity)

		bulk = append(bulk,
			tx.OrderItemProduct.Create().
				SetOrderID(orderID).
				SetOrderItemID(sourceOrderItemID).
				SetProductID(p.ProductID).
				SetOriginalOrderItemID(sourceOrderItemID).
				SetQuantity(qty).
				SetNillableProductCode(p.ProductCode).
				SetNillableRetailPrice(p.RetailPrice),
		)
	}

	if len(bulk) > 0 {
		if _, err := tx.OrderItemProduct.CreateBulk(bulk...).Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (r *orderItemProductRepository) syncFromSource(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	sourceOrderItemID int64,
	sourceProducts []*model.OrderItemProductDTO,
) error {

	// find derived order items
	var derivedOrderItemIDs []int64
	if err := tx.OrderItemProduct.
		Query().
		Where(
			orderitemproduct.OriginalOrderItemIDEQ(sourceOrderItemID),
			orderitemproduct.OrderItemIDNEQ(sourceOrderItemID),
		).
		GroupBy(orderitemproduct.FieldOrderItemID).
		Scan(ctx, &derivedOrderItemIDs); err != nil {
		return err
	}

	for _, derivedOID := range derivedOrderItemIDs {

		// delete derived products
		if _, err := tx.OrderItemProduct.Delete().
			Where(orderitemproduct.OrderItemIDEQ(derivedOID)).
			Exec(ctx); err != nil {
			return err
		}

		if len(sourceProducts) == 0 {
			continue
		}

		var bulk []*generated.OrderItemProductCreate

		for _, p := range sourceProducts {
			if p == nil || p.ProductID == 0 {
				continue
			}

			qty := r.normalizeQuantity(p.Quantity)

			bulk = append(bulk,
				tx.OrderItemProduct.Create().
					SetOrderID(orderID).
					SetOrderItemID(derivedOID).
					SetProductID(p.ProductID).
					SetOriginalOrderItemID(sourceOrderItemID).
					SetQuantity(qty).
					SetNillableProductCode(p.ProductCode).
					SetNillableRetailPrice(p.RetailPrice),
			)
		}

		if len(bulk) > 0 {
			if _, err := tx.OrderItemProduct.CreateBulk(bulk...).Save(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *orderItemProductRepository) GetProductsByOrderID(ctx context.Context, orderID int64) ([]*model.OrderItemProductDTO, error) {
	products, err := r.db.OrderItemProduct.
		Query().
		Where(
			orderitemproduct.OrderIDEQ(orderID),
			orderitemproduct.HasOrderItemWith(orderitem.DeletedAtIsNil()),
		).
		Select(
			orderitemproduct.FieldID,
			orderitemproduct.FieldOrderID,
			orderitemproduct.FieldOrderItemID,
			orderitemproduct.FieldOriginalOrderItemID,
			orderitemproduct.FieldProductID,
			orderitemproduct.FieldProductCode,
			orderitemproduct.FieldQuantity,
			orderitemproduct.FieldRetailPrice,
		).
		WithOrderItem(func(q *generated.OrderItemQuery) {
			q.Select(orderitem.FieldID, orderitem.FieldCode)
		}).
		WithProduct(func(q *generated.ProductQuery) {
			q.Select(product.FieldID, product.FieldCode, product.FieldName)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return nil, nil
	}

	out := make([]*model.OrderItemProductDTO, 0, len(products))
	for _, it := range products {
		dto := &model.OrderItemProductDTO{
			ID:                  it.ID,
			ProductCode:         it.ProductCode,
			ProductID:           it.ProductID,
			OrderItemID:         it.OrderItemID,
			OriginalOrderItemID: it.OriginalOrderItemID,
			OrderID:             it.OrderID,
			Quantity:            it.Quantity,
			RetailPrice:         it.RetailPrice,
		}
		if it.Edges.OrderItem != nil {
			dto.OrderItemCode = it.Edges.OrderItem.Code
		}
		if it.Edges.Product != nil {
			dto.ProductName = it.Edges.Product.Name
			if dto.ProductCode == nil {
				dto.ProductCode = it.Edges.Product.Code
			}
		}
		out = append(out, dto)
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

func (r *orderItemProductRepository) PrepareForRemake(
	ctx context.Context,
	items ...*model.OrderItemDTO,
) error {
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

	relations, err := r.db.OrderItemProduct.
		Query().
		Where(orderitemproduct.OrderItemIDIn(itemIDs...)).
		All(ctx)
	if err != nil {
		return err
	}

	for _, rel := range relations {
		dto, ok := itemIndex[rel.OrderItemID]
		if !ok {
			continue
		}

		mapped := mapper.MapAs[
			*generated.OrderItemProduct,
			*model.OrderItemProductDTO,
		](rel)

		cloneable := true
		mapped.IsCloneable = &cloneable

		dto.Products = append(dto.Products, mapped)
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
