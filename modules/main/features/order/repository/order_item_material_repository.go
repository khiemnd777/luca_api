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
			SetType(opts.materialType)

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

		// ONLY delete materials cloned from this source
		if _, err := tx.OrderItemMaterial.Delete().
			Where(
				orderitemmaterial.OrderItemIDEQ(derivedOID),
				orderitemmaterial.OriginalOrderItemIDEQ(sourceOrderItemID),
				orderitemmaterial.TypeEQ(opts.materialType),
			).
			Exec(ctx); err != nil {
			return err
		}

		if len(sourceMaterials) == 0 {
			continue
		}

		bulk := r.buildMaterialBulk(
			tx,
			orderID,
			derivedOID,
			sourceOrderItemID,
			sourceMaterials,
			opts,
		)

		if len(bulk) > 0 {
			if _, err := tx.OrderItemMaterial.CreateBulk(bulk...).Save(ctx); err != nil {
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
