package repository

import (
	"context"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitem"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemmaterial"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/utils"
)

func (r *orderItemMaterialRepository) PrepareConsumableMaterials(dto *model.OrderItemDTO) []*model.OrderItemMaterialDTO {
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
			ID:                  material.ID,
			MaterialID:          material.MaterialID,
			OrderItemID:         material.OrderItemID,
			OriginalOrderItemID: material.OriginalOrderItemID,
			OrderID:             material.OrderID,
			Quantity:            qty,
			RetailPrice:         material.RetailPrice,
			Type:                utils.Ptr("consumable"),
			IsCloneable:         material.IsCloneable,
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

func (r *orderItemMaterialRepository) GetConsumableTotalPriceByOrderID(ctx context.Context, tx *generated.Tx, orderID int64) (float64, error) {
	var c *generated.OrderItemMaterialClient
	if tx != nil {
		c = tx.OrderItemMaterial
	} else {
		c = r.db.OrderItemMaterial
	}
	materials, err := c.
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

func (r *orderItemMaterialRepository) PrepareConsumableForRemake(
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

	relations, err := r.db.OrderItemMaterial.
		Query().
		Where(
			orderitemmaterial.OrderItemIDIn(itemIDs...),
			orderitemmaterial.TypeEQ("consumable"),
		).
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
			*generated.OrderItemMaterial,
			*model.OrderItemMaterialDTO,
		](rel)

		cloneable := true
		mapped.IsCloneable = &cloneable

		dto.ConsumableMaterials = append(dto.ConsumableMaterials, mapped)
	}

	return nil
}

func (r *orderItemMaterialRepository) syncConsumableFromDerived(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	sourceOrderItemID int64,
	materials []*model.OrderItemMaterialDTO,
) error {

	if _, err := tx.OrderItemMaterial.Delete().
		Where(
			orderitemmaterial.OrderItemIDEQ(sourceOrderItemID),
			orderitemmaterial.OriginalOrderItemIDEQ(sourceOrderItemID),
			orderitemmaterial.TypeEQ("consumable"),
		).
		Exec(ctx); err != nil {
		return err
	}

	if len(materials) == 0 {
		return nil
	}

	bulk := r.buildMaterialBulk(
		tx,
		orderID,
		sourceOrderItemID,
		sourceOrderItemID,
		materials,
		materialBulkOptions{
			materialType:    "consumable",
			withRetailPrice: true,
		},
	)

	if len(bulk) > 0 {
		if _, err := tx.OrderItemMaterial.CreateBulk(bulk...).Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (r *orderItemMaterialRepository) syncConsumableFromSource(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	sourceOrderItemID int64,
	sourceMaterials []*model.OrderItemMaterialDTO,
) error {
	return r.syncFromSource(
		ctx,
		tx,
		orderID,
		sourceOrderItemID,
		sourceMaterials,
		materialBulkOptions{materialType: "consumable", withRetailPrice: true},
	)
}

func (r *orderItemMaterialRepository) replaceConsumableCurrent(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	orderItemID int64,
	materials []*model.OrderItemMaterialDTO,
) error {

	// Delete ALL consumable rows of CURRENT order item (FULL STATE)
	if _, err := tx.OrderItemMaterial.Delete().
		Where(
			orderitemmaterial.OrderItemIDEQ(orderItemID),
			orderitemmaterial.TypeEQ("consumable"),
		).
		Exec(ctx); err != nil {
		return err
	}

	if len(materials) == 0 {
		return nil
	}

	bulk := make([]*generated.OrderItemMaterialCreate, 0, len(materials))

	for _, m := range materials {
		if m == nil || m.MaterialID == 0 {
			continue
		}

		qty := r.normalizeQuantity(m.Quantity)

		origOID := orderItemID
		if m.OriginalOrderItemID != nil && *m.OriginalOrderItemID != 0 {
			origOID = *m.OriginalOrderItemID
		}

		c := tx.OrderItemMaterial.Create().
			SetOrderID(orderID).
			SetOrderItemID(orderItemID).
			SetOriginalOrderItemID(origOID).
			SetMaterialID(m.MaterialID).
			SetQuantity(qty).
			SetType("consumable").
			SetNillableRetailPrice(m.RetailPrice).
			SetNillableStatus(m.Status).
			SetNillableIsCloneable(m.IsCloneable)

		// Optional fields if your Ent schema has them:
		// c.SetNillableMaterialCode(m.MaterialCode)
		// c.SetNillableMaterialName(m.MaterialName)
		// c.SetNillableOrderItemCode(m.OrderItemCode)

		bulk = append(bulk, c)
	}

	if len(bulk) == 0 {
		return nil
	}

	_, err := tx.OrderItemMaterial.CreateBulk(bulk...).Save(ctx)
	return err
}

func (r *orderItemMaterialRepository) SyncConsumable(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	orderItemID int64,
	materials []*model.OrderItemMaterialDTO,
) ([]*model.OrderItemMaterialDTO, error) {

	logger.Debug("SyncConsumableV2: start",
		"orderItemID", orderItemID,
		"inputCount", len(materials),
	)

	current := make([]*model.OrderItemMaterialDTO, 0, len(materials))
	cloneToParent := make(map[int64][]*model.OrderItemMaterialDTO)
	cloneToChildren := make([]*model.OrderItemMaterialDTO, 0)

	for i, m := range materials {
		if m == nil || m.MaterialID == 0 {
			continue
		}

		logger.Debug("SyncConsumableV2: input material",
			"index", i,
			"id", m.ID,
			"materialID", m.MaterialID,
			"orderID", m.OrderID,
			"orderItemID", m.OrderItemID,
			"originalOrderItemID", m.OriginalOrderItemID,
			"quantity", m.Quantity,
			"retailPrice", m.RetailPrice,
			"type", m.Type,
			"status", m.Status,
			"isCloneable", m.IsCloneable,
			"isCloneableValue",
			func() any {
				if m.IsCloneable == nil {
					return nil
				}
				return *m.IsCloneable
			}(),
		)

		current = append(current, m)

		isCloneable := m.IsCloneable != nil && *m.IsCloneable

		if isCloneable {
			if m.OriginalOrderItemID != nil && *m.OriginalOrderItemID != orderItemID {
				parentOID := *m.OriginalOrderItemID
				cloneToParent[parentOID] = append(cloneToParent[parentOID], m)
			}
		} else {
			cloneToChildren = append(cloneToChildren, m)
		}
	}

	logger.Debug("SyncConsumableV2: partition",
		"current", len(current),
		"cloneUpGroups", len(cloneToParent),
		"cloneDown", len(cloneToChildren),
	)

	if err := r.replaceConsumableCurrent(
		ctx, tx, orderID, orderItemID, current,
	); err != nil {
		return nil, err
	}

	for parentOID, items := range cloneToParent {
		if parentOID == orderItemID {
			continue
		}

		if err := r.syncConsumableFromDerived(
			ctx,
			tx,
			orderID,
			parentOID,
			items,
		); err != nil {
			return nil, err
		}

		if err := r.syncConsumableFromSource(
			ctx,
			tx,
			orderID,
			parentOID,
			items,
		); err != nil {
			return nil, err
		}
	}

	if len(cloneToChildren) > 0 {
		if err := r.syncConsumableFromSource(
			ctx,
			tx,
			orderID,
			orderItemID,
			cloneToChildren,
		); err != nil {
			return nil, err
		}
	}

	rows, err := tx.OrderItemMaterial.
		Query().
		Where(
			orderitemmaterial.OrderItemIDEQ(orderItemID),
			orderitemmaterial.TypeEQ("consumable"),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*model.OrderItemMaterialDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out,
			mapper.MapAs[*generated.OrderItemMaterial, *model.OrderItemMaterialDTO](r),
		)
	}

	logger.Debug("SyncConsumableV2: done",
		"orderItemID", orderItemID,
		"finalCount", len(out),
	)

	return out, nil
}
