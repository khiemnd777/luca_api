package repository

import (
	"context"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemmaterial"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type OrderItemMaterialRepository interface {
	// Consumable
	PrepareConsumableMaterials(dto *model.OrderItemDTO) []*model.OrderItemMaterialDTO
	CalculateConsumableTotalPrice(materials []*model.OrderItemMaterialDTO) *float64
	GetConsumableTotalPriceByOrderItemID(ctx context.Context, orderItemID int64) (float64, error)
	GetConsumableTotalPriceByOrderID(ctx context.Context, tx *generated.Tx, orderID int64) (float64, error)
	PrepareConsumableForRemake(
		ctx context.Context,
		items ...*model.OrderItemDTO,
	) error
	LoadConsumable(ctx context.Context, items ...*model.OrderItemDTO) error

	// Loaner
	GetLoanerMaterials(ctx context.Context, query table.TableQuery) (table.TableListResult[model.OrderItemMaterialDTO], error)
	PrepareLoanerForRemake(
		ctx context.Context,
		items ...*model.OrderItemDTO,
	) error
	PrepareLoanerMaterials(dto *model.OrderItemDTO) []*model.OrderItemMaterialDTO
	LoadLoaner(ctx context.Context, items ...*model.OrderItemDTO) error

	SyncConsumable(
		ctx context.Context,
		tx *generated.Tx,
		orderID int64,
		orderItemID int64,
		materials []*model.OrderItemMaterialDTO,
	) ([]*model.OrderItemMaterialDTO, error)

	PrepareLoanerForCreate(materials []*model.OrderItemMaterialDTO) []*model.OrderItemMaterialDTO

	SyncLoaner(
		ctx context.Context,
		tx *generated.Tx,
		orderID int64,
		orderItemID int64,
		materials []*model.OrderItemMaterialDTO,
	) ([]*model.OrderItemMaterialDTO, error)
}

type orderItemMaterialRepository struct {
	db *generated.Client
}

func NewOrderItemMaterialRepository(db *generated.Client) OrderItemMaterialRepository {
	return &orderItemMaterialRepository{db: db}
}

type materialBulkOptions struct {
	materialType    string
	withRetailPrice bool
	withStatus      bool
}

func (r *orderItemMaterialRepository) buildMaterialBulk(
	tx *generated.Tx,
	orderID int64,
	orderItemID int64,
	originalOrderItemID int64,
	materials []*model.OrderItemMaterialDTO,
	opts materialBulkOptions,
) []*generated.OrderItemMaterialCreate {
	if len(materials) == 0 {
		return nil
	}

	bulk := make([]*generated.OrderItemMaterialCreate, 0, len(materials))
	for _, m := range materials {
		if m == nil || m.MaterialID == 0 {
			continue
		}

		qty := r.normalizeQuantity(m.Quantity)
		create := tx.OrderItemMaterial.Create().
			SetOrderID(orderID).
			SetOrderItemID(orderItemID).
			SetOriginalOrderItemID(originalOrderItemID).
			SetMaterialID(m.MaterialID).
			SetQuantity(qty).
			SetType(opts.materialType).
			SetNillableNote(m.Note)

		if opts.withRetailPrice {
			create.SetNillableRetailPrice(m.RetailPrice)
		}
		if opts.withStatus {
			create.SetNillableStatus(m.Status)
		}

		bulk = append(bulk, create)
	}

	return bulk
}

func (r *orderItemMaterialRepository) syncFromSource(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	sourceOrderItemID int64,
	sourceMaterials []*model.OrderItemMaterialDTO,
	opts materialBulkOptions,
) error {

	// 1. Find derived order items
	var derivedOrderItemIDs []int64
	if err := tx.OrderItemMaterial.
		Query().
		Where(
			orderitemmaterial.OriginalOrderItemIDEQ(sourceOrderItemID),
			orderitemmaterial.OrderItemIDNEQ(sourceOrderItemID),
			orderitemmaterial.TypeEQ(opts.materialType),
		).
		GroupBy(orderitemmaterial.FieldOrderItemID).
		Scan(ctx, &derivedOrderItemIDs); err != nil {
		return err
	}

	for _, derivedOID := range derivedOrderItemIDs {

		// 2. Load existing derived clones of THIS source
		existing, err := tx.OrderItemMaterial.
			Query().
			Where(
				orderitemmaterial.OrderItemIDEQ(derivedOID),
				orderitemmaterial.OriginalOrderItemIDEQ(sourceOrderItemID),
				orderitemmaterial.TypeEQ(opts.materialType),
			).
			All(ctx)
		if err != nil {
			return err
		}

		existingByMaterialID := map[int]*generated.OrderItemMaterial{}
		for _, row := range existing {
			existingByMaterialID[row.MaterialID] = row
		}

		seen := map[int]struct{}{}

		// 3. MERGE (UPSERT)
		for _, m := range sourceMaterials {
			if m == nil || m.MaterialID == 0 {
				continue
			}

			qty := r.normalizeQuantity(m.Quantity)
			seen[m.MaterialID] = struct{}{}

			if row, ok := existingByMaterialID[m.MaterialID]; ok {
				// UPDATE business fields only
				upd := tx.OrderItemMaterial.
					UpdateOne(row).
					SetQuantity(qty)

				if opts.withRetailPrice {
					upd.SetNillableRetailPrice(m.RetailPrice)
				}
				if opts.withStatus {
					upd.SetNillableStatus(m.Status)
				}

				if _, err := upd.Save(ctx); err != nil {
					return err
				}
			} else {
				// INSERT NEW CLONE
				create := tx.OrderItemMaterial.Create().
					SetOrderID(orderID).
					SetOrderItemID(derivedOID).
					SetOriginalOrderItemID(sourceOrderItemID).
					SetMaterialID(m.MaterialID).
					SetQuantity(qty).
					SetType(opts.materialType).
					SetIsCloneable(true).
					SetNillableNote(m.Note)

				if opts.withRetailPrice {
					create.SetNillableRetailPrice(m.RetailPrice)
				}
				if opts.withStatus {
					create.SetNillableStatus(m.Status)
				}

				if _, err := create.Save(ctx); err != nil {
					return err
				}
			}
		}

		// 4. DELETE removed clones
		for _, row := range existing {
			if _, ok := seen[row.MaterialID]; ok {
				continue
			}
			if err := tx.OrderItemMaterial.DeleteOne(row).Exec(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *orderItemMaterialRepository) normalizeQuantity(quantity int) int {
	if quantity <= 0 {
		return 1
	}
	return quantity
}
