package repository

import (
	"context"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/material"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitem"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemmaterial"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/utils"
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

func (r *orderItemMaterialRepository) PrepareLoanerMaterials(dto *model.OrderItemDTO) []*model.OrderItemMaterialDTO {
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
			ID:                  material.ID,
			MaterialID:          material.MaterialID,
			OrderItemID:         material.OrderItemID,
			OriginalOrderItemID: material.OriginalOrderItemID,
			OrderID:             material.OrderID,
			Quantity:            qty,
			Type:                utils.Ptr("loaner"),
			Status:              material.Status,
			IsCloneable:         material.IsCloneable,
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
		if _, err := tx.OrderItemMaterial.Delete().
			Where(
				orderitemmaterial.OrderItemIDEQ(derivedOID),
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

func (r *orderItemMaterialRepository) syncConsumableFromDerived(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	derivedOrderItemID int64,
	sourceOrderItemID int64,
	materials []*model.OrderItemMaterialDTO,
) error {
	if _, err := tx.OrderItemMaterial.Delete().
		Where(
			orderitemmaterial.OrderItemIDEQ(sourceOrderItemID),
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
		materialBulkOptions{materialType: "consumable", withRetailPrice: true},
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

func (r *orderItemMaterialRepository) deleteSourceConsumableCascade(
	ctx context.Context,
	tx *generated.Tx,
	sourceOrderItemID int64,
	materialID int,
) error {
	_, err := tx.OrderItemMaterial.Delete().
		Where(
			orderitemmaterial.OriginalOrderItemIDEQ(sourceOrderItemID),
			orderitemmaterial.MaterialIDEQ(materialID),
			orderitemmaterial.TypeEQ("consumable"),
		).
		Exec(ctx)

	return err
}

func (r *orderItemMaterialRepository) SyncConsumable(
	ctx context.Context,
	tx *generated.Tx,
	orderID,
	orderItemID int64,
	materials []*model.OrderItemMaterialDTO,
) ([]*model.OrderItemMaterialDTO, error) {
	existing, err := tx.OrderItemMaterial.
		Query().
		Where(
			orderitemmaterial.OrderItemIDEQ(orderItemID),
			orderitemmaterial.TypeEQ("consumable"),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	existingByMaterialID := make(map[int]*generated.OrderItemMaterial, len(existing))
	for _, e := range existing {
		existingByMaterialID[e.MaterialID] = e
	}

	inputByMaterialID := make(map[int]*model.OrderItemMaterialDTO, len(materials))
	sourceMaterials := make([]*model.OrderItemMaterialDTO, 0)
	derivedBySource := make(map[int64][]*model.OrderItemMaterialDTO)

	for _, m := range materials {
		if m == nil || m.MaterialID == 0 {
			continue
		}

		inputByMaterialID[m.MaterialID] = m

		if m.OriginalOrderItemID != nil && *m.OriginalOrderItemID != orderItemID {
			derivedBySource[*m.OriginalOrderItemID] =
				append(derivedBySource[*m.OriginalOrderItemID], m)
		} else {
			sourceMaterials = append(sourceMaterials, m)
		}
	}

	for sourceOID, items := range derivedBySource {
		if err := r.syncConsumableFromDerived(
			ctx,
			tx,
			orderID,
			orderItemID,
			sourceOID,
			items,
		); err != nil {
			return nil, err
		}
	}

	if len(sourceMaterials) > 0 {
		if err := r.syncConsumableFromSource(
			ctx,
			tx,
			orderID,
			orderItemID,
			sourceMaterials,
		); err != nil {
			return nil, err
		}
	}

	for materialID, row := range existingByMaterialID {
		if _, ok := inputByMaterialID[materialID]; ok {
			continue
		}

		isSource := row.OriginalOrderItemID != nil &&
			*row.OriginalOrderItemID == row.OrderItemID

		if isSource {
			if err := r.deleteSourceConsumableCascade(
				ctx,
				tx,
				row.OrderItemID,
				materialID,
			); err != nil {
				return nil, err
			}
		} else {
			if row.OriginalOrderItemID != nil {
				if err := r.syncConsumableFromDerived(
					ctx,
					tx,
					orderID,
					orderItemID,
					*row.OriginalOrderItemID,
					nil,
				); err != nil {
					return nil, err
				}
			}
		}
	}

	finalRows, err := tx.OrderItemMaterial.
		Query().
		Where(
			orderitemmaterial.OrderItemIDEQ(orderItemID),
			orderitemmaterial.TypeEQ("consumable"),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*model.OrderItemMaterialDTO, 0, len(finalRows))
	for _, row := range finalRows {
		out = append(out,
			mapper.MapAs[*generated.OrderItemMaterial, *model.OrderItemMaterialDTO](row),
		)
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
	existing, err := tx.OrderItemMaterial.
		Query().
		Where(
			orderitemmaterial.OrderItemIDEQ(orderItemID),
			orderitemmaterial.TypeEQ("loaner"),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	existingByMaterialID := make(map[int]*generated.OrderItemMaterial, len(existing))
	for _, e := range existing {
		existingByMaterialID[e.MaterialID] = e
	}

	inputByMaterialID := make(map[int]*model.OrderItemMaterialDTO, len(materials))
	sourceMaterials := make([]*model.OrderItemMaterialDTO, 0)
	derivedBySource := make(map[int64][]*model.OrderItemMaterialDTO)

	for _, m := range materials {
		if m == nil || m.MaterialID == 0 {
			continue
		}

		inputByMaterialID[m.MaterialID] = m

		if m.OriginalOrderItemID != nil && *m.OriginalOrderItemID != orderItemID {
			derivedBySource[*m.OriginalOrderItemID] =
				append(derivedBySource[*m.OriginalOrderItemID], m)
		} else {
			sourceMaterials = append(sourceMaterials, m)
		}
	}

	for sourceOID, items := range derivedBySource {
		if err := r.syncLoanerFromDerived(
			ctx,
			tx,
			orderID,
			orderItemID,
			sourceOID,
			items,
		); err != nil {
			return nil, err
		}
	}

	if len(sourceMaterials) > 0 {
		if err := r.syncLoanerFromSource(
			ctx,
			tx,
			orderID,
			orderItemID,
			sourceMaterials,
		); err != nil {
			return nil, err
		}
	}

	for materialID, row := range existingByMaterialID {
		if _, ok := inputByMaterialID[materialID]; ok {
			continue
		}

		isSource := row.OriginalOrderItemID != nil &&
			*row.OriginalOrderItemID == row.OrderItemID

		if isSource {
			if err := r.deleteSourceLoanerCascade(
				ctx,
				tx,
				row.OrderItemID,
				materialID,
			); err != nil {
				return nil, err
			}
		} else {
			if row.OriginalOrderItemID != nil {
				if err := r.syncLoanerFromDerived(
					ctx,
					tx,
					orderID,
					orderItemID,
					*row.OriginalOrderItemID,
					nil,
				); err != nil {
					return nil, err
				}
			}
		}
	}

	finalRows, err := tx.OrderItemMaterial.
		Query().
		Where(
			orderitemmaterial.OrderItemIDEQ(orderItemID),
			orderitemmaterial.TypeEQ("loaner"),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*model.OrderItemMaterialDTO, 0, len(finalRows))
	for _, row := range finalRows {
		out = append(out,
			mapper.MapAs[*generated.OrderItemMaterial, *model.OrderItemMaterialDTO](row),
		)
	}

	return out, nil
}

func (r *orderItemMaterialRepository) syncLoanerFromDerived(
	ctx context.Context,
	tx *generated.Tx,
	orderID int64,
	derivedOrderItemID int64,
	sourceOrderItemID int64,
	materials []*model.OrderItemMaterialDTO,
) error {
	// delete all loaner materials of source
	if _, err := tx.OrderItemMaterial.Delete().
		Where(
			orderitemmaterial.OrderItemIDEQ(sourceOrderItemID),
			orderitemmaterial.TypeEQ("loaner"),
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
		materialBulkOptions{materialType: "loaner", withStatus: true},
	)
	if len(bulk) > 0 {
		if _, err := tx.OrderItemMaterial.CreateBulk(bulk...).Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (r *orderItemMaterialRepository) syncLoanerFromSource(
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
		materialBulkOptions{materialType: "loaner", withStatus: true},
	)
}

func (r *orderItemMaterialRepository) deleteSourceLoanerCascade(
	ctx context.Context,
	tx *generated.Tx,
	sourceOrderItemID int64,
	materialID int,
) error {
	_, err := tx.OrderItemMaterial.Delete().
		Where(
			orderitemmaterial.OriginalOrderItemIDEQ(sourceOrderItemID),
			orderitemmaterial.MaterialIDEQ(materialID),
			orderitemmaterial.TypeEQ("loaner"),
		).
		Exec(ctx)

	return err
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

func (r *orderItemMaterialRepository) PrepareLoanerForRemake(
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
			orderitemmaterial.TypeEQ("loaner"),
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

		dto.LoanerMaterials = append(dto.LoanerMaterials, mapped)
	}

	return nil
}

func (r *orderItemMaterialRepository) normalizeQuantity(quantity int) int {
	if quantity <= 0 {
		return 1
	}
	return quantity
}
