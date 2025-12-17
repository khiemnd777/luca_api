package repository

import (
	"context"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitem"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemmaterial"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type OrderItemMaterialRepository interface {
	// Consumable
	CollectConsumableMaterials(dto *model.OrderItemDTO) []*model.OrderItemMaterialDTO
	CalculateConsumableTotalPrice(materials []*model.OrderItemMaterialDTO) *float64
	GetConsumableTotalPriceByOrderItemID(ctx context.Context, orderItemID int64) (float64, error)
	GetConsumableTotalPriceByOrderID(ctx context.Context, orderID int64) (float64, error)
	LoadConsumable(ctx context.Context, items ...*model.OrderItemDTO) error

	// Loaner
	CollectLoanerMaterials(dto *model.OrderItemDTO) []*model.OrderItemMaterialDTO
	LoadLoaner(ctx context.Context, items ...*model.OrderItemDTO) error

	Sync(
		ctx context.Context,
		tx *generated.Tx,
		orderID,
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

func (r *orderItemMaterialRepository) CollectConsumableMaterials(dto *model.OrderItemDTO) []*model.OrderItemMaterialDTO {
	if dto == nil || len(dto.ConsumableMaterials) == 0 {
		return nil
	}

	out := make([]*model.OrderItemMaterialDTO, 0, len(dto.ConsumableMaterials))
	seen := make(map[int]struct{}, len(dto.ConsumableMaterials))

	for _, material := range dto.ConsumableMaterials {
		if material == nil || material.MaterialID == 0 {
			continue
		}
		if _, ok := seen[material.MaterialID]; ok {
			continue
		}
		seen[material.MaterialID] = struct{}{}

		qty := r.normalizeQuantity(material.Quantity)
		out = append(out, &model.OrderItemMaterialDTO{
			ID:          material.ID,
			MaterialID:  material.MaterialID,
			OrderItemID: material.OrderItemID,
			OrderID:     material.OrderID,
			Quantity:    qty,
			Type:        utils.Ptr("consumable"),
		})
	}

	return out
}

func (r *orderItemMaterialRepository) CalculateConsumableTotalPrice(consumableMaterials []*model.OrderItemMaterialDTO) *float64 {
	var total float64
	hasPrice := false

	for _, consumableMaterial := range consumableMaterials {
		if consumableMaterial == nil || consumableMaterial.RetailPrice == nil {
			continue
		}
		qty := r.normalizeQuantity(consumableMaterial.Quantity)
		total += *consumableMaterial.RetailPrice * float64(qty)
		hasPrice = true
	}

	if !hasPrice {
		return nil
	}

	return &total
}

func (r *orderItemMaterialRepository) GetConsumableTotalPriceByOrderItemID(ctx context.Context, orderItemID int64) (float64, error) {
	materials, err := r.db.OrderItemMaterial.
		Query().
		Where(
			orderitemmaterial.OrderItemIDEQ(orderItemID),
			orderitemmaterial.TypeEQ("consumable"),
		).
		Select(orderitemmaterial.FieldQuantity, orderitemmaterial.FieldRetailPrice).
		All(ctx)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, material := range materials {
		if material == nil || material.RetailPrice == nil {
			continue
		}
		qty := r.normalizeQuantity(material.Quantity)
		total += *material.RetailPrice * float64(qty)
	}

	return total, nil
}

func (r *orderItemMaterialRepository) GetConsumableTotalPriceByOrderID(ctx context.Context, orderID int64) (float64, error) {
	materials, err := r.db.OrderItemMaterial.
		Query().
		Where(
			orderitemmaterial.TypeEQ("consumable"),
			orderitemmaterial.OrderIDEQ(orderID),
			orderitemmaterial.HasOrderItemWith(orderitem.DeletedAtIsNil()),
		).
		Select(orderitemmaterial.FieldQuantity, orderitemmaterial.FieldRetailPrice).
		All(ctx)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, material := range materials {
		if material == nil || material.RetailPrice == nil {
			continue
		}
		qty := r.normalizeQuantity(material.Quantity)
		total += *material.RetailPrice * float64(qty)
	}

	return total, nil
}

func (r *orderItemMaterialRepository) CollectLoanerMaterials(dto *model.OrderItemDTO) []*model.OrderItemMaterialDTO {
	if dto == nil || len(dto.LoanerMaterials) == 0 {
		return nil
	}

	out := make([]*model.OrderItemMaterialDTO, 0, len(dto.LoanerMaterials))
	seen := make(map[int]struct{}, len(dto.LoanerMaterials))

	for _, material := range dto.LoanerMaterials {
		if material == nil || material.MaterialID == 0 {
			continue
		}
		if _, ok := seen[material.MaterialID]; ok {
			continue
		}
		seen[material.MaterialID] = struct{}{}

		qty := r.normalizeQuantity(material.Quantity)
		out = append(out, &model.OrderItemMaterialDTO{
			ID:          material.ID,
			MaterialID:  material.MaterialID,
			OrderItemID: material.OrderItemID,
			OrderID:     material.OrderID,
			Quantity:    qty,
			Type:        utils.Ptr("loaner"),
			Status:      material.Status,
		})
	}

	return out
}

func (r *orderItemMaterialRepository) Sync(
	ctx context.Context,
	tx *generated.Tx,
	orderID,
	orderItemID int64,
	materials []*model.OrderItemMaterialDTO,
) ([]*model.OrderItemMaterialDTO, error) {
	_, err := tx.OrderItemMaterial.Delete().
		Where(orderitemmaterial.OrderItemIDEQ(orderItemID)).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	if len(materials) == 0 {
		return nil, nil
	}

	bulk := make([]*generated.OrderItemMaterialCreate, 0, len(materials))
	for _, material := range materials {
		if material == nil || material.MaterialID == 0 {
			continue
		}

		qty := r.normalizeQuantity(material.Quantity)
		create := tx.OrderItemMaterial.Create().
			SetOrderID(orderID).
			SetOrderItemID(orderItemID).
			SetMaterialID(material.MaterialID).
			SetQuantity(qty).
			SetNillableType(material.Type).
			SetNillableStatus(material.Status)

		bulk = append(bulk, create)
	}

	if len(bulk) == 0 {
		return nil, nil
	}

	created, err := tx.OrderItemMaterial.CreateBulk(bulk...).Save(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*model.OrderItemMaterialDTO, 0, len(created))
	for _, it := range created {
		out = append(out, mapper.MapAs[*generated.OrderItemMaterial, *model.OrderItemMaterialDTO](it))
	}

	return out, nil
}

func (r *orderItemMaterialRepository) LoadConsumable(ctx context.Context, items ...*model.OrderItemDTO) error {
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

	relations, err := r.db.OrderItemMaterial.Query().
		Where(orderitemmaterial.OrderItemIDIn(itemIDs...)).
		All(ctx)
	if err != nil {
		return err
	}

	for _, rel := range relations {
		if dto, ok := itemIndex[rel.OrderItemID]; ok {
			dto.ConsumableMaterials = append(dto.ConsumableMaterials, mapper.MapAs[*generated.OrderItemMaterial, *model.OrderItemMaterialDTO](rel))
		}
	}

	return nil
}

func (r *orderItemMaterialRepository) LoadLoaner(ctx context.Context, items ...*model.OrderItemDTO) error {
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

	relations, err := r.db.OrderItemMaterial.Query().
		Where(orderitemmaterial.OrderItemIDIn(itemIDs...)).
		All(ctx)
	if err != nil {
		return err
	}

	for _, rel := range relations {
		if dto, ok := itemIndex[rel.OrderItemID]; ok {
			dto.LoanerMaterials = append(dto.LoanerMaterials, mapper.MapAs[*generated.OrderItemMaterial, *model.OrderItemMaterialDTO](rel))
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
