package repository

import (
	"context"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/material"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitem"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemmaterial"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type OrderItemMaterialRepository interface {
	// Consumable
	CollectConsumableMaterials(dto *model.OrderItemDTO) []*model.OrderItemMaterialDTO
	CalculateConsumableTotalPrice(materials []*model.OrderItemMaterialDTO) *float64
	GetConsumableTotalPriceByOrderItemID(ctx context.Context, orderItemID int64) (float64, error)
	GetConsumableTotalPriceByOrderID(ctx context.Context, orderID int64) (float64, error)
	LoadConsumable(ctx context.Context, items ...*model.OrderItemDTO) error

	// Loaner
	GetLoanerMaterials(ctx context.Context, query table.TableQuery) (table.TableListResult[model.OrderItemMaterialDTO], error)
	CollectLoanerMaterials(dto *model.OrderItemDTO) []*model.OrderItemMaterialDTO
	LoadLoaner(ctx context.Context, items ...*model.OrderItemDTO) error

	SyncConsumable(
		ctx context.Context,
		tx *generated.Tx,
		orderID,
		orderItemID int64,
		materials []*model.OrderItemMaterialDTO,
	) ([]*model.OrderItemMaterialDTO, error)

	PrepareLoanerForCreate(materials []*model.OrderItemMaterialDTO) []*model.OrderItemMaterialDTO

	SyncLoaner(
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
	if dto == nil {
		return nil
	}

	if len(dto.ConsumableMaterials) == 0 {
		return nil
	}

	out := make([]*model.OrderItemMaterialDTO, 0, len(dto.ConsumableMaterials))
	seen := make(map[int]struct{}, len(dto.ConsumableMaterials))
	invalidCount := 0
	duplicateCount := 0

	for _, material := range dto.ConsumableMaterials {
		if material == nil || material.MaterialID == 0 {
			invalidCount++
			continue
		}
		if _, ok := seen[material.MaterialID]; ok {
			duplicateCount++
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
			RetailPrice: material.RetailPrice,
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

func (r *orderItemMaterialRepository) GetLoanerMaterials(
	ctx context.Context,
	query table.TableQuery,
) (table.TableListResult[model.OrderItemMaterialDTO], error) {
	base := r.db.OrderItemMaterial.
		Query().
		Where(
			orderitemmaterial.TypeEQ("loaner"),
			orderitemmaterial.StatusEQ("on_loan"),
		)

	list, err := table.TableListV2(
		ctx,
		base,
		query,
		orderitemmaterial.Table,
		orderitemmaterial.FieldID,
		orderitemmaterial.FieldID,
		func(q *generated.OrderItemMaterialQuery) *generated.OrderItemMaterialQuery {
			return q.
				Select(
					orderitemmaterial.FieldID,
					orderitemmaterial.FieldMaterialCode,
					orderitemmaterial.FieldMaterialID,
					orderitemmaterial.FieldOrderItemID,
					orderitemmaterial.FieldOrderID,
					orderitemmaterial.FieldQuantity,
					orderitemmaterial.FieldType,
					orderitemmaterial.FieldStatus,
					orderitemmaterial.FieldRetailPrice,
				).
				WithOrderItem(func(oq *generated.OrderItemQuery) {
					oq.Select(orderitem.FieldCode)
				}).
				WithMaterial(func(mq *generated.MaterialQuery) {
					mq.Select(material.FieldName)
				})
		},
		func(src []*generated.OrderItemMaterial) []*model.OrderItemMaterialDTO {
			out := make([]*model.OrderItemMaterialDTO, 0, len(src))
			for _, item := range src {
				if item == nil {
					continue
				}
				dto := mapper.MapAs[*generated.OrderItemMaterial, *model.OrderItemMaterialDTO](item)
				if item.Edges.OrderItem != nil {
					dto.OrderItemCode = item.Edges.OrderItem.Code
				}
				if item.Edges.Material != nil {
					dto.MaterialName = item.Edges.Material.Name
				}
				out = append(out, dto)
			}
			return out
		},
	)
	if err != nil {
		var zero table.TableListResult[model.OrderItemMaterialDTO]
		return zero, err
	}

	return list, nil
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

func (r *orderItemMaterialRepository) PrepareLoanerForCreate(materials []*model.OrderItemMaterialDTO) []*model.OrderItemMaterialDTO {
	if len(materials) == 0 {
		return materials
	}

	status := utils.Ptr("on_loan")
	for _, material := range materials {
		if material == nil {
			continue
		}
		material.Status = status
	}

	return materials
}

func (r *orderItemMaterialRepository) SyncConsumable(
	ctx context.Context,
	tx *generated.Tx,
	orderID,
	orderItemID int64,
	materials []*model.OrderItemMaterialDTO,
) ([]*model.OrderItemMaterialDTO, error) {
	_, err := tx.OrderItemMaterial.Delete().
		Where(
			orderitemmaterial.OrderItemIDEQ(orderItemID),
			orderitemmaterial.TypeEQ("consumable"),
		).
		Exec(ctx)
	if err != nil {
		logger.Debug("Sync order item material delete failed", "err", err, "orderItemID", orderItemID, "orderID", orderID)
		return nil, err
	}

	if len(materials) == 0 {
		logger.Debug("Sync order item material skipped, no materials", "orderItemID", orderItemID, "orderID", orderID)
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
			SetType("consumable").
			SetNillableRetailPrice(material.RetailPrice)

		bulk = append(bulk, create)
	}

	if len(bulk) == 0 {
		logger.Debug("Sync order item material skipped, no valid materials", "orderItemID", orderItemID, "orderID", orderID, "materialsLen", len(materials))
		return nil, nil
	}

	created, err := tx.OrderItemMaterial.CreateBulk(bulk...).Save(ctx)
	if err != nil {
		logger.Debug("Sync order item material create failed", "err", err, "orderItemID", orderItemID, "orderID", orderID, "bulkLen", len(bulk))
		return nil, err
	}

	out := make([]*model.OrderItemMaterialDTO, 0, len(created))
	for _, it := range created {
		out = append(out, mapper.MapAs[*generated.OrderItemMaterial, *model.OrderItemMaterialDTO](it))
	}

	return out, nil
}

func (r *orderItemMaterialRepository) SyncLoaner(
	ctx context.Context,
	tx *generated.Tx,
	orderID,
	orderItemID int64,
	materials []*model.OrderItemMaterialDTO,
) ([]*model.OrderItemMaterialDTO, error) {
	_, err := tx.OrderItemMaterial.Delete().
		Where(
			orderitemmaterial.OrderItemIDEQ(orderItemID),
			orderitemmaterial.TypeEQ("loaner"),
		).
		Exec(ctx)
	if err != nil {
		logger.Debug("Sync order item material delete failed", "err", err, "orderItemID", orderItemID, "orderID", orderID)
		return nil, err
	}

	if len(materials) == 0 {
		logger.Debug("Sync order item material skipped, no materials", "orderItemID", orderItemID, "orderID", orderID)
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
			SetType("loaner").
			SetNillableStatus(material.Status)

		bulk = append(bulk, create)
	}

	if len(bulk) == 0 {
		logger.Debug("Sync order item material skipped, no valid materials", "orderItemID", orderItemID, "orderID", orderID, "materialsLen", len(materials))
		return nil, nil
	}

	created, err := tx.OrderItemMaterial.CreateBulk(bulk...).Save(ctx)
	if err != nil {
		logger.Debug("Sync order item material create failed", "err", err, "orderItemID", orderItemID, "orderID", orderID, "bulkLen", len(bulk))
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
		Where(
			orderitemmaterial.OrderItemIDIn(itemIDs...),
			orderitemmaterial.TypeEQ("consumable"),
		).
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
		Where(
			orderitemmaterial.OrderItemIDIn(itemIDs...),
			orderitemmaterial.TypeEQ("loaner"),
		).
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
