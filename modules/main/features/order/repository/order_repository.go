package repository

import (
	"context"
	"fmt"
	"maps"
	"math"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/modules/main/features/promotion/contextbuilder"
	"github.com/khiemnd777/andy_api/modules/main/features/promotion/engine"
	promotionrepo "github.com/khiemnd777/andy_api/modules/main/features/promotion/repository"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/material"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/order"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitem"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemmaterial"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemprocess"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/orderitemproduct"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/product"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type OrderRepository interface {
	ExistsByCode(ctx context.Context, code string) (bool, error)
	GetByOrderIDAndOrderItemID(ctx context.Context, orderID, orderItemID int64) (*model.OrderDTO, error)
	UpdateStatus(ctx context.Context, orderItemProcessID int64, status string) (*model.OrderItemDTO, error)
	SyncPrice(ctx context.Context, orderID int64) (float64, error)
	GetAllOrderProducts(ctx context.Context, orderID int64) ([]*model.OrderItemProductDTO, error)
	GetAllOrderMaterials(ctx context.Context, orderID int64) ([]*model.OrderItemMaterialDTO, error)
	// -- general functions
	Create(ctx context.Context, userID int, input *model.OrderUpsertDTO) (*model.OrderDTO, error)
	Update(ctx context.Context, userID int, input *model.OrderUpsertDTO) (*model.OrderDTO, error)
	GetByID(ctx context.Context, id int64) (*model.OrderDTO, error)
	PrepareForRemakeByOrderID(
		ctx context.Context,
		orderID int64,
	) (*model.OrderDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.OrderDTO], error)
	GetOrdersBySectionID(ctx context.Context, sectionID int, query table.TableQuery) (table.TableListResult[model.OrderDTO], error)
	InProgressList(ctx context.Context, query table.TableQuery) (table.TableListResult[model.InProcessOrderDTO], error)
	NewestList(ctx context.Context, query table.TableQuery) (table.TableListResult[model.NewestOrderDTO], error)
	CompletedList(ctx context.Context, query table.TableQuery) (table.TableListResult[model.CompletedOrderDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.OrderDTO], error)
	Delete(ctx context.Context, id int64) error
}

type orderRepository struct {
	db                   *generated.Client
	deps                 *module.ModuleDeps[config.ModuleConfig]
	cfMgr                *customfields.Manager
	orderItemRepo        OrderItemRepository
	orderItemProcessRepo OrderItemProcessRepository
	orderItemProductRepo OrderItemProductRepository
	orderCodeRepo        OrderCodeRepository
	promotionRepo        promotionrepo.PromotionRepository
	promoengine          *engine.Engine
	promoctxbuilder      *contextbuilder.Builder
	promoguard           engine.PromotionGuard
}

func NewOrderRepository(
	db *generated.Client,
	deps *module.ModuleDeps[config.ModuleConfig],
	cfMgr *customfields.Manager,
) OrderRepository {
	orderItemProductRepo := NewOrderItemProductRepository(db)
	promotionRepo := promotionrepo.NewPromotionRepository(db, deps.DB)
	promoengine := engine.NewEngine(deps)
	promoctxbuilder := contextbuilder.NewBuilder(orderItemProductRepo)
	promoguard := engine.NewGuard(promotionRepo)

	return &orderRepository{
		db:                   db,
		deps:                 deps,
		cfMgr:                cfMgr,
		orderItemRepo:        NewOrderItemRepository(db, deps, cfMgr),
		orderItemProcessRepo: NewOrderItemProcessRepository(db, deps, cfMgr),
		orderCodeRepo:        NewOrderCodeRepository(db),
		orderItemProductRepo: orderItemProductRepo,
		promotionRepo:        promotionRepo,
		promoengine:          promoengine,
		promoctxbuilder:      promoctxbuilder,
		promoguard:           promoguard,
	}
}

func (r *orderRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	return r.db.Order.
		Query().
		Where(
			order.CodeEQ(code),
			order.DeletedAtIsNil(),
		).
		Exist(ctx)
}

func (r *orderRepository) GetByOrderIDAndOrderItemID(ctx context.Context, orderID, orderItemID int64) (*model.OrderDTO, error) {
	q := r.db.Order.Query().
		Where(
			order.ID(orderID),
			order.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Order, *model.OrderDTO](entity)

	// latest order item
	latest, err := r.orderItemRepo.GetByID(ctx, orderItemID)
	if err != nil {
		return nil, err
	}
	dto.LatestOrderItem = latest
	return dto, nil
}

func (r *orderRepository) SyncPrice(ctx context.Context, orderID int64) (float64, error) {
	return r.orderItemRepo.GetTotalPriceByOrderID(ctx, nil, orderID)
}

// -- helpers

func (r *orderRepository) createNewOrder(
	ctx context.Context,
	tx *generated.Tx,
	userID int,
	input *model.OrderUpsertDTO,
) (*model.OrderDTO, error) {

	dto := &input.DTO

	logger.Debug(
		"create order: dto fields",
		"clinic_id", utils.DerefInt(dto.ClinicID),
		"clinic_name", utils.DerefString(dto.ClinicName),
		"dentist_id", utils.DerefInt(dto.DentistID),
		"dentist_name", utils.DerefString(dto.DentistName),
		"patient_id", utils.DerefInt(dto.PatientID),
		"patient_name", utils.DerefString(dto.PatientName),
		"ref_user_name", utils.DerefString(dto.RefUserName),
	)

	q := tx.Order.Create().
		SetNillableCode(dto.Code).
		SetNillablePromotionCode(dto.PromotionCode).
		SetNillablePromotionCodeID(dto.PromotionCodeID).
		SetNillableClinicID(dto.ClinicID).
		SetNillableClinicName(dto.ClinicName).
		SetNillableDentistID(dto.DentistID).
		SetNillableDentistName(dto.DentistName).
		SetNillablePatientID(dto.PatientID).
		SetNillablePatientName(dto.PatientName).
		SetNillableRefUserID(dto.RefUserID).
		SetNillableRefUserName(dto.RefUserName)

	// custom fields
	if input.Collections != nil && len(*input.Collections) > 0 {
		if _, err := customfields.PrepareCustomFields(
			ctx, r.cfMgr, *input.Collections, dto.CustomFields, q, false,
		); err != nil {
			return nil, err
		}
	}

	// save order
	orderEnt, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	// map back
	out := mapper.MapAs[*generated.Order, *model.OrderDTO](orderEnt)

	// create first-latest order item
	loi := input.DTO.LatestOrderItemUpsert
	loi.DTO.OrderID = out.ID
	loi.DTO.CodeOriginal = out.Code

	latest, err := r.orderItemRepo.Create(ctx, tx, out, loi)
	if err != nil {
		return nil, err
	}

	out.LatestOrderItem = latest

	// reassign latest order item -> order as cache to appear them on the table
	lstStatus := utils.SafeGetStringPtr(latest.CustomFields, "status")
	lstPriority := utils.SafeGetStringPtr(latest.CustomFields, "priority")
	prdQty := utils.SafeGetIntPtr(latest.CustomFields, "quantity")
	dlrDate := utils.SafeGetDateTimePtr(latest.CustomFields, "delivery_date")
	rmkType := utils.SafeGetStringPtr(latest.CustomFields, "remake_type")
	rmkCount := latest.RemakeCount

	// total price
	totalPrice, err := r.orderItemRepo.GetTotalPriceByOrderID(ctx, tx, out.ID)
	if err != nil {
		return nil, err
	}
	prdTotalPrice := totalPrice

	discountAmount, promoSnapshot := r.buildPromotionSnapshot(ctx, out)
	if discountAmount > 0 {
		prdTotalPrice = math.Max(0, prdTotalPrice-discountAmount)
	}

	_, err = orderEnt.
		Update().
		SetNillableCodeLatest(latest.Code).
		SetNillableStatusLatest(lstStatus).
		SetNillablePriorityLatest(lstPriority).
		SetNillableQuantity(prdQty).
		SetTotalPrice(prdTotalPrice).
		SetNillableDeliveryDate(dlrDate).
		SetNillableRemakeType(rmkType).
		SetNillableRemakeCount(&rmkCount).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	// Assign latest ones to output
	out.CodeLatest = latest.Code
	out.StatusLatest = lstStatus
	out.PriorityLatest = lstPriority
	out.Quantity = prdQty
	out.TotalPrice = &prdTotalPrice
	out.DeliveryDate = dlrDate
	out.RemakeType = rmkType
	out.RemakeCount = &rmkCount

	processes, err := r.orderItemProcessRepo.GetProcessesByOrderItemID(ctx, tx, out.LatestOrderItem.ID)

	if err != nil {
		logger.Error(
			"failed to get processes by order item",
			"orderItemID", out.LatestOrderItem.ID,
			"err", err,
		)
		return nil, err
	}

	if len(processes) > 0 {
		stProc := processes[0]
		out.ProcessIDLatest = utils.Ptr(int(stProc.ID))
		out.ProcessNameLatest = stProc.ProcessName
		out.SectionNameLatest = stProc.SectionName
		out.LeaderIDLatest = stProc.LeaderID
		out.LeaderNameLatest = stProc.LeaderName
	}

	err = r.orderCodeRepo.ConfirmReservation(ctx, tx, *dto.Code)
	if err != nil {
		return nil, err
	}

	// relation
	// if err = relation.Upsert1(ctx, tx, "orders_customers", orderEnt, &input.DTO, out); err != nil {
	// 	return nil, err
	// }
	if err = relation.Upsert1(ctx, tx, "orders_clinics", orderEnt, &input.DTO, out); err != nil {
		return nil, err
	}
	if err = relation.Upsert1(ctx, tx, "orders_dentists", orderEnt, &input.DTO, out); err != nil {
		return nil, err
	}
	if err = relation.Upsert1(ctx, tx, "orders_patients", orderEnt, &input.DTO, out); err != nil {
		return nil, err
	}
	if err = relation.Upsert1(ctx, tx, "orders_ref_users", orderEnt, &input.DTO, out); err != nil {
		return nil, err
	}

	if promoSnapshot != nil {
		if err := r.promotionRepo.UpsertPromotionUsageFromSnapshot(
			ctx,
			tx,
			*out.PromotionCodeID,
			out.ID,
			out.RefUserID,
			promoSnapshot,
		); err != nil {
			return nil, err
		}
	}

	return out, nil
}

func (r *orderRepository) upsertExistingOrder(
	ctx context.Context,
	tx *generated.Tx,
	userID int,
	input *model.OrderUpsertDTO,
) (*model.OrderDTO, error) {

	dto := &input.DTO

	// Load order theo code
	orderEnt, err := r.db.Order.
		Query().
		Where(order.CodeEQ(*dto.Code), order.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	// UPDATE ORDER (custom fields + m2m + 1)
	up := tx.Order.UpdateOneID(orderEnt.ID).
		SetNillableCode(dto.Code)

	if input.Collections != nil && len(*input.Collections) > 0 {
		if _, err := customfields.PrepareCustomFields(
			ctx, r.cfMgr, *input.Collections, dto.CustomFields, up, false,
		); err != nil {
			return nil, err
		}
	}

	orderEnt, err = up.Save(ctx)
	if err != nil {
		return nil, err
	}

	out := mapper.MapAs[*generated.Order, *model.OrderDTO](orderEnt)

	loi := input.DTO.LatestOrderItemUpsert
	loi.DTO.OrderID = out.ID
	loi.DTO.CodeOriginal = out.Code

	latest, err := r.orderItemRepo.Create(ctx, tx, out, loi)
	if err != nil {
		return nil, err
	}

	out.LatestOrderItem = latest

	// reassign latest order item -> order as cache to appear them on the table
	lstStatus := utils.SafeGetStringPtr(latest.CustomFields, "status")
	lstPriority := utils.SafeGetStringPtr(latest.CustomFields, "priority")
	dlrDate := utils.SafeGetDateTimePtr(latest.CustomFields, "delivery_date")
	rmkType := utils.SafeGetStringPtr(latest.CustomFields, "remake_type")
	rmkCount := latest.RemakeCount

	totalPrice, err := r.orderItemRepo.GetTotalPriceByOrderID(ctx, tx, out.ID)
	if err != nil {
		return nil, err
	}
	prdTotalPrice := totalPrice

	discountAmount, promoSnapshot := r.buildPromotionSnapshot(ctx, out)
	if discountAmount > 0 {
		prdTotalPrice = math.Max(0, prdTotalPrice-discountAmount)
	}

	_, err = orderEnt.
		Update().
		SetNillableCodeLatest(latest.Code).
		SetNillableStatusLatest(lstStatus).
		SetNillablePriorityLatest(lstPriority).
		SetTotalPrice(prdTotalPrice).
		SetNillableDeliveryDate(dlrDate).
		SetNillableRemakeType(rmkType).
		SetNillableRemakeCount(&rmkCount).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	// Assign latest ones to output
	out.CodeLatest = latest.Code
	out.StatusLatest = lstStatus
	out.PriorityLatest = lstPriority
	out.TotalPrice = &prdTotalPrice
	out.DeliveryDate = dlrDate
	out.RemakeType = rmkType
	out.RemakeCount = &rmkCount

	// relations
	// if err := relation.Upsert1(ctx, tx, "orders_customers", orderEnt, &input.DTO, out); err != nil {
	// 	return nil, err
	// }
	if err = relation.Upsert1(ctx, tx, "orders_clinics", orderEnt, &input.DTO, out); err != nil {
		return nil, err
	}
	if err = relation.Upsert1(ctx, tx, "orders_dentists", orderEnt, &input.DTO, out); err != nil {
		return nil, err
	}
	if err = relation.Upsert1(ctx, tx, "orders_patients", orderEnt, &input.DTO, out); err != nil {
		return nil, err
	}
	if err = relation.Upsert1(ctx, tx, "orders_ref_users", orderEnt, &input.DTO, out); err != nil {
		return nil, err
	}

	if promoSnapshot != nil {
		if err := r.promotionRepo.UpsertPromotionUsageFromSnapshot(
			ctx,
			tx,
			*out.PromotionCodeID,
			out.ID,
			out.RefUserID,
			promoSnapshot,
		); err != nil {
			return nil, err
		}
	}

	return out, nil
}

// -- general functions
func (r *orderRepository) Create(ctx context.Context, userID int, input *model.OrderUpsertDTO) (*model.OrderDTO, error) {
	var err error

	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			logger.Error(fmt.Sprintf("[ERROR] %v", err))
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	dto := &input.DTO
	code := dto.Code

	exists, err := r.ExistsByCode(ctx, *code)
	if err != nil {
		return nil, err
	}

	var out *model.OrderDTO
	if exists {
		out, err = r.upsertExistingOrder(ctx, tx, userID, input)
		if err != nil {
			return nil, err
		}
	} else {
		out, err = r.createNewOrder(ctx, tx, userID, input)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func (r *orderRepository) Update(
	ctx context.Context,
	userID int,
	input *model.OrderUpsertDTO,
) (*model.OrderDTO, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	output := &input.DTO

	q := tx.Order.UpdateOneID(output.ID).
		SetNillableClinicID(output.ClinicID).
		SetNillableClinicName(output.ClinicName).
		SetNillablePromotionCode(output.PromotionCode).
		SetNillablePromotionCodeID(output.PromotionCodeID).
		SetNillableDentistID(output.DentistID).
		SetNillableDentistName(output.DentistName).
		SetNillablePatientID(output.PatientID).
		SetNillablePatientName(output.PatientName).
		SetNillableRefUserID(output.RefUserID).
		SetNillableRefUserName(output.RefUserName)

	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err = customfields.PrepareCustomFields(
			ctx,
			r.cfMgr,
			*input.Collections,
			output.CustomFields,
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

	output = mapper.MapAs[*generated.Order, *model.OrderDTO](entity)

	// ===== Update latest order item
	latest, err := r.orderItemRepo.Update(
		ctx,
		tx,
		output,
		input.DTO.LatestOrderItemUpsert,
	)
	if err != nil {
		return nil, err
	}

	isLatest, err := r.orderItemRepo.IsLatestIfOrderID(
		ctx,
		entity.ID,
		latest.ID,
	)
	if err != nil {
		return nil, err
	}

	output.LatestOrderItem = latest

	// ===== Update order cache fields & total price
	if isLatest {
		lstStatus := utils.SafeGetStringPtr(latest.CustomFields, "status")
		lstPriority := utils.SafeGetStringPtr(latest.CustomFields, "priority")
		dlrDate := utils.SafeGetDateTimePtr(latest.CustomFields, "delivery_date")
		rmkType := utils.SafeGetStringPtr(latest.CustomFields, "remake_type")
		rmkCount := latest.RemakeCount

		totalPrice, err := r.orderItemRepo.GetTotalPriceByOrderID(
			ctx,
			tx,
			output.ID,
		)
		if err != nil {
			return nil, err
		}

		discountAmount, promoSnapshot := r.buildPromotionSnapshot(ctx, output)

		if discountAmount > 0 {
			logger.Info(
				"apply promotion discount to order",
				"order_id", output.ID,
				"user_id", userID,
				"original_total_price", totalPrice,
				"discount_amount", discountAmount,
			)

			totalPrice = math.Max(0, totalPrice-discountAmount)
		} else {
			logger.Debug(
				"no promotion discount applied",
				"order_id", output.ID,
				"user_id", userID,
				"total_price", totalPrice,
			)
		}

		prdTotalPrice := totalPrice

		_, err = entity.
			Update().
			SetNillableCodeLatest(latest.Code).
			SetNillableStatusLatest(lstStatus).
			SetNillablePriorityLatest(lstPriority).
			SetTotalPrice(prdTotalPrice).
			SetNillableDeliveryDate(dlrDate).
			SetNillableRemakeType(rmkType).
			SetNillableRemakeCount(&rmkCount).
			Save(ctx)
		if err != nil {
			logger.Error(
				"failed to update order after applying promotion",
				"order_id", output.ID,
				"final_total_price", prdTotalPrice,
				"err", err,
			)
			return nil, err
		}

		logger.Debug(
			"order updated with final price",
			"order_id", output.ID,
			"final_total_price", prdTotalPrice,
			"status_latest", lstStatus,
			"priority_latest", lstPriority,
		)

		// ===== Assign latest values to output
		output.CodeLatest = latest.Code
		output.StatusLatest = lstStatus
		output.PriorityLatest = lstPriority
		output.TotalPrice = &prdTotalPrice
		output.DeliveryDate = dlrDate
		output.RemakeType = rmkType
		output.RemakeCount = &rmkCount

		// ===== Persist promotion usage snapshot
		if promoSnapshot != nil {
			logger.Info(
				"persist promotion usage snapshot",
				"order_id", output.ID,
				"user_id", userID,
				"promo_code", promoSnapshot.PromoCode,
				"discount_amount", promoSnapshot.DiscountAmount,
				"applied_conditions", promoSnapshot.AppliedConditions,
			)

			if err := r.promotionRepo.UpsertPromotionUsageFromSnapshot(
				ctx,
				tx,
				*output.PromotionCodeID,
				output.ID,
				output.RefUserID,
				promoSnapshot,
			); err != nil {
				logger.Error(
					"failed to persist promotion usage snapshot",
					"order_id", output.ID,
					"user_id", userID,
					"promo_code", promoSnapshot.PromoCode,
					"err", err,
				)
				return nil, err
			}
		}
	}

	// ===== Relations
	if err = relation.Upsert1(ctx, tx, "orders_clinics", entity, &input.DTO, output); err != nil {
		return nil, err
	}
	if err = relation.Upsert1(ctx, tx, "orders_dentists", entity, &input.DTO, output); err != nil {
		return nil, err
	}
	if err = relation.Upsert1(ctx, tx, "orders_patients", entity, &input.DTO, output); err != nil {
		return nil, err
	}
	if err = relation.Upsert1(ctx, tx, "orders_ref_users", entity, &input.DTO, output); err != nil {
		return nil, err
	}

	return output, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, orderItemProcessID int64, status string) (*model.OrderItemDTO, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	_, err = r.orderItemProcessRepo.UpdateStatus(ctx, tx, orderItemProcessID, status)
	if err != nil {
		return nil, err
	}

	// Get oip from memory to ensure CF == latest status, because not yet Committed to db
	updatedOIP, err := tx.OrderItemProcess.
		Query().
		Where(orderitemprocess.IDEQ(orderItemProcessID)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	if updatedOIP.OrderID == nil {
		err = fmt.Errorf("OrderID is nil after updating process")
		return nil, err
	}

	processes, err := r.orderItemProcessRepo.GetProcessesByOrderItemID(ctx, tx, updatedOIP.OrderItemID)
	if err != nil {
		return nil, err
	}

	orderDTO, err := r.recalculateOrderStatusByProcesses(ctx, tx, *updatedOIP.OrderID, updatedOIP.OrderItemID, processes)
	if err != nil {
		return nil, err
	}

	return orderDTO, nil
}

func (r *orderRepository) recalculateOrderStatusByProcesses(
	ctx context.Context,
	tx *generated.Tx,
	orderID,
	orderItemID int64,
	processes []*model.OrderItemProcessDTO,
) (*model.OrderItemDTO, error) {

	if len(processes) == 0 {
		return nil, fmt.Errorf("no processes found for order %d", orderItemID)
	}

	allWaiting := true
	allCompleted := true
	anyInProgress := false

	for _, p := range processes {
		status := utils.SafeGetString(p.CustomFields, "status")

		switch status {
		case "waiting":
		case "completed":
			allWaiting = false
		case "in_progress", "qc", "rework":
			allWaiting = false
			allCompleted = false
			anyInProgress = true
		default:
			allWaiting = false
			allCompleted = false
		}

		if status != "waiting" {
			allWaiting = false
		}
		if status != "completed" {
			allCompleted = false
		}
	}

	var orderStatus string

	switch {
	case allWaiting:
		orderStatus = "received"

	case anyInProgress:
		orderStatus = "in_progress"

	case allCompleted:
		orderStatus = "completed"

	default:
		orderStatus = "in_progress"
	}

	oi, err := tx.OrderItem.Query().
		Where(orderitem.IDEQ(orderItemID)).
		First(ctx)
	if err != nil {
		return nil, err
	}

	cf := maps.Clone(oi.CustomFields)
	cf["status"] = orderStatus

	updated, err := tx.OrderItem.
		UpdateOne(oi).
		SetCustomFields(cf).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.OrderItem, *model.OrderItemDTO](updated)
	dto.TotalPrice = updated.TotalPrice

	tx.Order.UpdateOneID(orderID).
		SetNillableStatusLatest(&orderStatus).
		Save(ctx)

	return dto, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*model.OrderDTO, error) {
	q := r.db.Order.Query().
		Where(
			order.ID(id),
			order.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Order, *model.OrderDTO](entity)

	// latest order item
	latest, err := r.orderItemRepo.GetLatestByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}
	dto.LatestOrderItem = latest
	return dto, nil
}

func (r *orderRepository) PrepareForRemakeByOrderID(
	ctx context.Context,
	orderID int64,
) (*model.OrderDTO, error) {

	entity, err := r.db.Order.
		Query().
		Where(
			order.ID(orderID),
			order.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[
		*generated.Order,
		*model.OrderDTO,
	](entity)

	latestItem, err := r.orderItemRepo.
		PrepareLatestForRemakeByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	dto.LatestOrderItem = latestItem

	return dto, nil
}

func (r *orderRepository) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.OrderDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Order.Query().
			Where(order.DeletedAtIsNil()),
		query,
		order.Table,
		order.FieldID,
		order.FieldID,
		func(src []*generated.Order) []*model.OrderDTO {
			return mapper.MapListAs[*generated.Order, *model.OrderDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.OrderDTO]
		return zero, err
	}
	return list, nil
}

func (r *orderRepository) GetOrdersBySectionID(
	ctx context.Context,
	sectionID int,
	query table.TableQuery,
) (table.TableListResult[model.OrderDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Order.Query().
			Where(
				order.DeletedAtIsNil(),
				order.HasItemsWith(
					orderitem.HasProcessesWith(
						orderitemprocess.SectionIDEQ(sectionID),
					),
				),
			),
		query,
		order.Table,
		order.FieldID,
		order.FieldID,
		func(src []*generated.Order) []*model.OrderDTO {
			return mapper.MapListAs[*generated.Order, *model.OrderDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.OrderDTO]
		return zero, err
	}
	return list, nil
}

func (r *orderRepository) InProgressList(ctx context.Context, query table.TableQuery) (table.TableListResult[model.InProcessOrderDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Order.Query().
			Where(
				order.DeletedAtIsNil(),
				order.StatusLatestEQ("in_progress"),
			),
		query,
		order.Table,
		order.FieldID,
		order.FieldDeliveryDate,
		func(src []*generated.Order) []*model.InProcessOrderDTO {
			out := make([]*model.InProcessOrderDTO, 0, len(src))
			for _, item := range src {
				out = append(out, &model.InProcessOrderDTO{
					ID:                item.ID,
					Code:              item.Code,
					CodeLatest:        item.CodeLatest,
					DeliveryDate:      item.DeliveryDate,
					TotalPrice:        item.TotalPrice,
					ProcessNameLatest: item.ProcessNameLatest,
					StatusLatest:      item.StatusLatest,
					PriorityLatest:    item.PriorityLatest,
				})
			}
			return out
		},
	)
	if err != nil {
		var zero table.TableListResult[model.InProcessOrderDTO]
		return zero, err
	}
	return list, nil
}

func (r *orderRepository) NewestList(ctx context.Context, query table.TableQuery) (table.TableListResult[model.NewestOrderDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Order.Query().
			Where(
				order.DeletedAtIsNil(),
				order.StatusLatestEQ("received"),
			),
		query,
		order.Table,
		order.FieldID,
		order.FieldCreatedAt,
		func(src []*generated.Order) []*model.NewestOrderDTO {
			out := make([]*model.NewestOrderDTO, 0, len(src))
			for _, item := range src {
				out = append(out, &model.NewestOrderDTO{
					ID:             item.ID,
					Code:           item.Code,
					CodeLatest:     item.CodeLatest,
					CreatedAt:      item.CreatedAt,
					StatusLatest:   item.StatusLatest,
					PriorityLatest: item.PriorityLatest,
				})
			}
			return out
		},
	)
	if err != nil {
		var zero table.TableListResult[model.NewestOrderDTO]
		return zero, err
	}
	return list, nil
}

func (r *orderRepository) CompletedList(ctx context.Context, query table.TableQuery) (table.TableListResult[model.CompletedOrderDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Order.Query().
			Where(
				order.DeletedAtIsNil(),
				order.StatusLatestEQ("completed"),
			),
		query,
		order.Table,
		order.FieldID,
		order.FieldUpdatedAt,
		func(src []*generated.Order) []*model.CompletedOrderDTO {
			out := make([]*model.CompletedOrderDTO, 0, len(src))
			for _, item := range src {
				out = append(out, &model.CompletedOrderDTO{
					ID:             item.ID,
					Code:           item.Code,
					CodeLatest:     item.CodeLatest,
					CreatedAt:      item.CreatedAt,
					StatusLatest:   item.StatusLatest,
					PriorityLatest: item.PriorityLatest,
				})
			}
			return out
		},
	)
	if err != nil {
		var zero table.TableListResult[model.CompletedOrderDTO]
		return zero, err
	}
	return list, nil
}

func (r *orderRepository) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.OrderDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Order.Query().
			Where(order.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(order.FieldCode),
		},
		query,
		order.Table,
		order.FieldID,
		order.FieldID,
		order.Or,
		func(src []*generated.Order) []*model.OrderDTO {
			return mapper.MapListAs[*generated.Order, *model.OrderDTO](src)
		},
	)
}

func (r *orderRepository) Delete(ctx context.Context, id int64) error {
	hasItems, err := r.db.OrderItem.
		Query().
		Where(
			orderitem.OrderID(id),
			orderitem.DeletedAtIsNil(),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if hasItems {
		return fmt.Errorf("cannot delete order %d because it still has order items", id)
	}
	return r.db.Order.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}

func (r *orderRepository) GetAllOrderProducts(ctx context.Context, orderID int64) ([]*model.OrderItemProductDTO, error) {
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
			Note:                it.Note,
			RetailPrice:         it.RetailPrice,
			TeethPosition:       it.TeethPosition,
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

func (r *orderRepository) GetAllOrderMaterials(ctx context.Context, orderID int64) ([]*model.OrderItemMaterialDTO, error) {
	materials, err := r.db.OrderItemMaterial.
		Query().
		Where(
			orderitemmaterial.OrderIDEQ(orderID),
			orderitemmaterial.HasOrderItemWith(orderitem.DeletedAtIsNil()),
		).
		Select(
			orderitemmaterial.FieldID,
			orderitemmaterial.FieldOrderID,
			orderitemmaterial.FieldOrderItemID,
			orderitemmaterial.FieldMaterialID,
			orderitemmaterial.FieldMaterialCode,
			orderitemmaterial.FieldQuantity,
			orderitemmaterial.FieldRetailPrice,
			orderitemmaterial.FieldType,
		).
		WithOrderItem(func(q *generated.OrderItemQuery) {
			q.Select(orderitem.FieldID, orderitem.FieldCode)
		}).
		WithMaterial(func(q *generated.MaterialQuery) {
			q.Select(material.FieldID, material.FieldCode, material.FieldName)
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(materials) == 0 {
		return nil, nil
	}

	out := make([]*model.OrderItemMaterialDTO, 0, len(materials))
	for _, it := range materials {
		dto := &model.OrderItemMaterialDTO{
			ID:           it.ID,
			MaterialCode: it.MaterialCode,
			MaterialID:   it.MaterialID,
			OrderItemID:  it.OrderItemID,
			OrderID:      it.OrderID,
			Quantity:     it.Quantity,
			Note:         it.Note,
			RetailPrice:  it.RetailPrice,
			Type:         it.Type,
		}
		if it.Edges.OrderItem != nil {
			dto.OrderItemCode = it.Edges.OrderItem.Code
		}
		if it.Edges.Material != nil {
			dto.MaterialName = it.Edges.Material.Name
			if dto.MaterialCode == nil {
				dto.MaterialCode = it.Edges.Material.Code
			}
		}
		out = append(out, dto)
	}

	return out, nil
}
