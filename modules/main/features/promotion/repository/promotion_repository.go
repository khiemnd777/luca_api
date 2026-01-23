package repository

import (
	"context"
	"database/sql"
	"errors"
	"math"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotioncode"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotionusage"
	"github.com/khiemnd777/andy_api/shared/mapper"
)

type PromotionRepository interface {
	GetPromotionCodesInUsageByOrderID(ctx context.Context, orderID int64) ([]model.PromotionCodeDTO, error)
	ExistsUsageByOrderID(ctx context.Context, promoCodeID int, orderID int64) (bool, error)
	GetByCode(ctx context.Context, code string) (*generated.PromotionCode, error)
	CountUsageByUser(ctx context.Context, promoCodeID, userID int) (int, error)
	CountTotalUsage(ctx context.Context, promoCodeID int) (int, error)
	CreateUsage(ctx context.Context, tx *generated.Tx, usage *generated.PromotionUsage) error
	LockPromotionForUpdate(ctx context.Context, tx *generated.Tx, promoCodeID int) (*generated.PromotionCode, error)
	CreatePromotionUsageFromSnapshot(ctx context.Context, tx *generated.Tx, promoCodeID int, orderID int64, userID int, snapshot *model.PromotionSnapshot) error
	UpsertPromotionUsageFromSnapshot(
		ctx context.Context,
		tx *generated.Tx,
		promoCodeID int,
		orderID int64,
		userID int,
		snapshot *model.PromotionSnapshot,
	) error
}

type promotionRepository struct {
	db    *generated.Client
	sqlDB *sql.DB
}

func NewPromotionRepository(db *generated.Client, sqlDB *sql.DB) PromotionRepository {
	return &promotionRepository{db: db, sqlDB: sqlDB}
}

func (r *promotionRepository) GetByCode(ctx context.Context, code string) (*generated.PromotionCode, error) {
	return r.db.PromotionCode.
		Query().
		Where(promotioncode.CodeEQ(code)).
		WithScopes().
		WithConditions().
		Only(ctx)
}

func (r *promotionRepository) CountUsageByUserV1(ctx context.Context, promoCodeID, userID int) (int, error) {
	return r.db.PromotionUsage.
		Query().
		Where(
			promotionusage.PromoCodeID(promoCodeID),
			promotionusage.UserID(userID),
		).
		Count(ctx)
}

func (r *promotionRepository) CountUsageByUser(
	ctx context.Context,
	promoCodeID int,
	userID int,
) (int, error) {
	const q = `
SELECT COUNT(DISTINCT order_id)
FROM promotion_usages
WHERE promo_code_id = $1
  AND user_id = $2
`
	var count int
	if err := r.sqlDB.QueryRowContext(ctx, q, promoCodeID, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *promotionRepository) CountTotalUsageV1(ctx context.Context, promoCodeID int) (int, error) {
	return r.db.PromotionUsage.
		Query().
		Where(promotionusage.PromoCodeID(promoCodeID)).
		Count(ctx)
}

func (r *promotionRepository) CountTotalUsage(
	ctx context.Context,
	promoCodeID int,
) (int, error) {
	const q = `
SELECT COUNT(DISTINCT order_id)
FROM promotion_usages
WHERE promo_code_id = $1
`
	var count int
	if err := r.sqlDB.QueryRowContext(ctx, q, promoCodeID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *promotionRepository) ExistsUsageByOrderID(
	ctx context.Context,
	promoCodeID int,
	orderID int64,
) (bool, error) {
	const q = `
SELECT EXISTS (
  SELECT 1
  FROM promotion_usages
  WHERE promo_code_id = $1
    AND order_id = $2
  LIMIT 1
)
`
	var exists bool
	if err := r.sqlDB.QueryRowContext(ctx, q, promoCodeID, orderID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *promotionRepository) GetPromotionCodesInUsageByOrderID(
	ctx context.Context,
	orderID int64,
) ([]model.PromotionCodeDTO, error) {
	promos, err := r.db.PromotionCode.
		Query().
		Where(promotioncode.HasUsagesWith(promotionusage.OrderID(orderID))).
		WithScopes().
		WithConditions().
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]model.PromotionCodeDTO, 0, len(promos))
	for _, promo := range promos {
		if promo == nil {
			continue
		}
		dto := mapPromotionCodeDTO(promo)
		if dto == nil {
			continue
		}
		out = append(out, *dto)
	}
	return out, nil
}

func (r *promotionRepository) CreateUsage(ctx context.Context, tx *generated.Tx, usage *generated.PromotionUsage) error {
	if tx == nil {
		return errors.New("transaction is required")
	}
	if usage == nil {
		return errors.New("usage is required")
	}

	q := tx.PromotionUsage.
		Create().
		SetPromoCodeID(usage.PromoCodeID).
		SetOrderID(usage.OrderID).
		SetUserID(usage.UserID).
		SetDiscountAmount(usage.DiscountAmount)
	if !usage.UsedAt.IsZero() {
		q.SetUsedAt(usage.UsedAt)
	}
	_, err := q.Save(ctx)
	return err
}

func (r *promotionRepository) LockPromotionForUpdate(
	ctx context.Context,
	tx *generated.Tx,
	promoCodeID int,
) (*generated.PromotionCode, error) {

	if tx == nil {
		return nil, errors.New("transaction is required")
	}

	const q = `
SELECT
  id,
  code,
  discount_type,
  discount_value,
  max_discount_amount,
  min_order_value,
  total_usage_limit,
  usage_per_user,
  start_at,
  end_at,
  is_active
FROM promotion_codes
WHERE id = $1
FOR UPDATE
`

	rows, err := tx.QueryContext(ctx, q, promoCodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	var p generated.PromotionCode
	if err := rows.Scan(
		&p.ID,
		&p.Code,
		&p.DiscountType,
		&p.DiscountValue,
		&p.MaxDiscountAmount,
		&p.MinOrderValue,
		&p.TotalUsageLimit,
		&p.UsagePerUser,
		&p.StartAt,
		&p.EndAt,
		&p.IsActive,
	); err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *promotionRepository) UpsertPromotionUsageFromSnapshot(
	ctx context.Context,
	tx *generated.Tx,
	promoCodeID int,
	orderID int64,
	userID int,
	snapshot *model.PromotionSnapshot,
) error {

	if snapshot == nil {
		return errors.New("snapshot is required")
	}

	discountAmount := int(math.Round(snapshot.DiscountAmount))

	var (
		client *generated.Client
	)
	if tx != nil {
		client = tx.Client()
	} else {
		client = r.db
	}

	// TRY FIND EXISTING USAGE
	existing, err := client.PromotionUsage.
		Query().
		Where(
			promotionusage.OrderIDEQ(orderID),
			promotionusage.PromoCodeIDEQ(promoCodeID),
			promotionusage.DiscountAmountEQ(discountAmount),
		).
		Only(ctx)

	if err == nil {
		// UPDATE
		_, err = client.PromotionUsage.
			UpdateOne(existing).
			SetUserID(userID).
			SetPromoCode(snapshot.PromoCode).
			SetDiscountType(snapshot.DiscountType).
			SetDiscountValue(snapshot.DiscountValue).
			SetIsRemake(snapshot.IsRemake).
			SetRemakeCount(snapshot.RemakeCount).
			SetAppliedConditions(snapshot.AppliedConditions).
			SetAppliedAt(snapshot.AppliedAt).
			SetUsedAt(snapshot.AppliedAt).
			Save(ctx)

		return err
	}

	if !generated.IsNotFound(err) {
		// real error
		return err
	}

	// CREATE
	_, err = client.PromotionUsage.
		Create().
		SetPromoCodeID(promoCodeID).
		SetOrderID(orderID).
		SetUserID(userID).
		SetPromoCode(snapshot.PromoCode).
		SetDiscountType(snapshot.DiscountType).
		SetDiscountValue(snapshot.DiscountValue).
		SetDiscountAmount(discountAmount).
		SetIsRemake(snapshot.IsRemake).
		SetRemakeCount(snapshot.RemakeCount).
		SetAppliedConditions(snapshot.AppliedConditions).
		SetAppliedAt(snapshot.AppliedAt).
		SetUsedAt(snapshot.AppliedAt).
		Save(ctx)

	return err
}

func (r *promotionRepository) CreatePromotionUsageFromSnapshot(
	ctx context.Context,
	tx *generated.Tx,
	promoCodeID int,
	orderID int64,
	userID int,
	snapshot *model.PromotionSnapshot,
) error {
	if snapshot == nil {
		return errors.New("snapshot is required")
	}

	discountAmount := int(math.Round(snapshot.DiscountAmount))

	q := r.db.PromotionUsage.Create()
	if tx != nil {
		q = tx.PromotionUsage.Create()
	}

	_, err := q.
		SetPromoCodeID(promoCodeID).
		SetOrderID(orderID).
		SetUserID(userID).
		SetPromoCode(snapshot.PromoCode).
		SetDiscountType(snapshot.DiscountType).
		SetDiscountValue(snapshot.DiscountValue).
		SetDiscountAmount(discountAmount).
		SetIsRemake(snapshot.IsRemake).
		SetRemakeCount(snapshot.RemakeCount).
		SetAppliedConditions(snapshot.AppliedConditions).
		SetAppliedAt(snapshot.AppliedAt).
		SetUsedAt(snapshot.AppliedAt).
		Save(ctx)
	return err
}

func mapPromotionCodeDTOFromInput(
	promo *generated.PromotionCode,
	scopes []model.PromotionScopeInput,
	conditions []model.PromotionConditionInput,
) *model.PromotionCodeDTO {
	dto := mapper.MapAs[*generated.PromotionCode, *model.PromotionCodeDTO](promo)
	dto.Scopes = scopes
	dto.Conditions = conditions
	return dto
}

func mapPromotionCodeDTO(promo *generated.PromotionCode) *model.PromotionCodeDTO {
	dto := mapper.MapAs[*generated.PromotionCode, *model.PromotionCodeDTO](promo)
	dto.Scopes = mapPromotionScopes(promo.Edges.Scopes)
	dto.Conditions = mapPromotionConditions(promo.Edges.Conditions)
	return dto
}

func mapPromotionScopes(scopes []*generated.PromotionScope) []model.PromotionScopeInput {
	if len(scopes) == 0 {
		if scopes == nil {
			return nil
		}
		return []model.PromotionScopeInput{}
	}
	out := make([]model.PromotionScopeInput, 0, len(scopes))
	for _, scope := range scopes {
		if scope == nil {
			continue
		}
		out = append(out, model.PromotionScopeInput{
			ScopeType:  scope.ScopeType,
			ScopeValue: scope.ScopeValue,
		})
	}
	return out
}

func mapPromotionConditions(conditions []*generated.PromotionCondition) []model.PromotionConditionInput {
	if len(conditions) == 0 {
		if conditions == nil {
			return nil
		}
		return []model.PromotionConditionInput{}
	}
	out := make([]model.PromotionConditionInput, 0, len(conditions))
	for _, condition := range conditions {
		if condition == nil {
			continue
		}
		out = append(out, model.PromotionConditionInput{
			ConditionType:  condition.ConditionType,
			ConditionValue: condition.ConditionValue,
		})
	}
	return out
}
