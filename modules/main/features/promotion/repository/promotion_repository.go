package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotioncode"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotionusage"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type PromotionRepository interface {
	// ===== Checkout / Business =====
	GetByCode(ctx context.Context, code string) (*generated.PromotionCode, error)
	CountUsageByUser(ctx context.Context, promoCodeID, userID int) (int, error)
	CountTotalUsage(ctx context.Context, promoCodeID int) (int, error)
	CreateUsage(ctx context.Context, tx *generated.Tx, usage *generated.PromotionUsage) error
	LockPromotionForUpdate(ctx context.Context, tx *generated.Tx, promoCodeID int) (*generated.PromotionCode, error)
	CreateOrderPromotionSnapshot(ctx context.Context, orderID int64, snapshot *model.PromotionSnapshot) error

	// ===== Admin CRUD =====
	CreatePromotion(ctx context.Context, input *model.CreatePromotionInput) (*generated.PromotionCode, error)
	UpdatePromotion(ctx context.Context, id int, input *model.UpdatePromotionInput) (*generated.PromotionCode, error)
	DeletePromotion(ctx context.Context, id int) error
	GetPromotionByID(ctx context.Context, id int) (*generated.PromotionCode, error)
	ListPromotions(ctx context.Context, query table.TableQuery) (table.TableListResult[model.PromotionCodeDTO], error)
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

func (r *promotionRepository) CountUsageByUser(ctx context.Context, promoCodeID, userID int) (int, error) {
	return r.db.PromotionUsage.
		Query().
		Where(
			promotionusage.PromoCodeID(promoCodeID),
			promotionusage.UserID(userID),
		).
		Count(ctx)
}

func (r *promotionRepository) CountTotalUsage(ctx context.Context, promoCodeID int) (int, error) {
	return r.db.PromotionUsage.
		Query().
		Where(promotionusage.PromoCodeID(promoCodeID)).
		Count(ctx)
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

func (r *promotionRepository) CreateOrderPromotionSnapshot(
	ctx context.Context,
	orderID int64,
	snapshot *model.PromotionSnapshot,
) error {
	if r.sqlDB == nil {
		return errors.New("sql db is required")
	}
	if snapshot == nil {
		return errors.New("snapshot is required")
	}

	appliedJSON, err := json.Marshal(snapshot.AppliedConditions)
	if err != nil {
		return err
	}

	discountAmount := int(math.Round(snapshot.DiscountAmount))

	const insertSQL = `
INSERT INTO order_promotions(
  order_id,
  promo_code,
  discount_type,
  discount_value,
  discount_amount,
  is_remake,
  remake_count,
  applied_conditions,
  applied_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

	_, err = r.sqlDB.ExecContext(
		ctx,
		insertSQL,
		orderID,
		snapshot.PromoCode,
		snapshot.DiscountType,
		snapshot.DiscountValue,
		discountAmount,
		snapshot.IsRemake,
		snapshot.RemakeCount,
		appliedJSON,
		snapshot.AppliedAt,
	)
	return err
}

func (r *promotionRepository) CreatePromotion(
	ctx context.Context,
	input *model.CreatePromotionInput,
) (*generated.PromotionCode, error) {

	if input == nil {
		return nil, errors.New("input is required")
	}

	q := r.db.PromotionCode.
		Create().
		SetCode(input.Code).
		SetDiscountType(input.DiscountType).
		SetDiscountValue(input.DiscountValue).
		SetIsActive(input.IsActive)

	if input.MaxDiscountAmount != nil {
		q.SetMaxDiscountAmount(*input.MaxDiscountAmount)
	}
	if input.MinOrderValue != nil {
		q.SetMinOrderValue(*input.MinOrderValue)
	}
	if input.TotalUsageLimit != nil {
		q.SetTotalUsageLimit(*input.TotalUsageLimit)
	}
	if input.UsagePerUser != nil {
		q.SetUsagePerUser(*input.UsagePerUser)
	}
	if input.StartAt != nil {
		q.SetStartAt(*input.StartAt)
	}
	if input.EndAt != nil {
		q.SetEndAt(*input.EndAt)
	}

	return q.Save(ctx)
}

func (r *promotionRepository) UpdatePromotion(
	ctx context.Context,
	id int,
	input *model.UpdatePromotionInput,
) (*generated.PromotionCode, error) {

	if input == nil {
		return nil, errors.New("input is required")
	}

	q := r.db.PromotionCode.
		UpdateOneID(id)

	q.SetDiscountType(input.DiscountType)

	q.SetDiscountValue(input.DiscountValue)

	if input.MaxDiscountAmount != nil {
		q.SetMaxDiscountAmount(*input.MaxDiscountAmount)
	}
	if input.MinOrderValue != nil {
		q.SetMinOrderValue(*input.MinOrderValue)
	}
	if input.TotalUsageLimit != nil {
		q.SetTotalUsageLimit(*input.TotalUsageLimit)
	}
	if input.UsagePerUser != nil {
		q.SetUsagePerUser(*input.UsagePerUser)
	}
	if input.StartAt != nil {
		q.SetStartAt(*input.StartAt)
	}
	if input.EndAt != nil {
		q.SetEndAt(*input.EndAt)
	}
	if input.IsActive != nil {
		q.SetIsActive(*input.IsActive)
	}

	return q.Save(ctx)
}

func (r *promotionRepository) DeletePromotion(ctx context.Context, id int) error {
	return r.db.PromotionCode.
		DeleteOneID(id).
		Exec(ctx)
}

func (r *promotionRepository) GetPromotionByID(
	ctx context.Context,
	id int,
) (*generated.PromotionCode, error) {

	return r.db.PromotionCode.
		Query().
		Where(promotioncode.ID(id)).
		WithScopes().
		WithConditions().
		Only(ctx)
}

func (r *promotionRepository) ListPromotions(
	ctx context.Context,
	query table.TableQuery,
) (table.TableListResult[model.PromotionCodeDTO], error) {
	if query.Direction == "" {
		query.Direction = "desc"
	}

	base := r.db.PromotionCode.Query()
	list, err := table.TableListV2(
		ctx,
		base,
		query,
		promotioncode.Table,
		promotioncode.FieldID,
		promotioncode.FieldCreatedAt,
		func(q *generated.PromotionCodeQuery) *generated.PromotionCodeQuery {
			return q
		},
		func(src []*generated.PromotionCode) []*model.PromotionCodeDTO {
			return mapper.MapListAs[*generated.PromotionCode, *model.PromotionCodeDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.PromotionCodeDTO]
		return zero, err
	}
	return list, nil
}
