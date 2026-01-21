package repository

import (
	"context"
	"database/sql"
	"errors"
	"math"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	promotionmodel "github.com/khiemnd777/andy_api/modules/main/features/promotion/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotioncode"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotioncondition"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotionscope"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotionusage"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
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
	CreatePromotionUsageFromSnapshot(ctx context.Context, promoCodeID int, orderID int, userID int, snapshot *model.PromotionSnapshot) error

	// ===== Admin CRUD =====
	CreatePromotion(ctx context.Context, input *model.CreatePromotionInput) (*model.PromotionCodeDTO, error)
	UpdatePromotion(ctx context.Context, id int, input *model.UpdatePromotionInput) (*model.PromotionCodeDTO, error)
	DeletePromotion(ctx context.Context, id int) error
	GetPromotionByID(ctx context.Context, id int) (*model.PromotionCodeDTO, error)
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

func (r *promotionRepository) CreatePromotionUsageFromSnapshot(
	ctx context.Context,
	promoCodeID int,
	orderID int,
	userID int,
	snapshot *model.PromotionSnapshot,
) error {
	if snapshot == nil {
		return errors.New("snapshot is required")
	}

	discountAmount := int(math.Round(snapshot.DiscountAmount))

	_, err := r.db.PromotionUsage.
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

func (r *promotionRepository) CreatePromotion(
	ctx context.Context,
	input *model.CreatePromotionInput,
) (*model.PromotionCodeDTO, error) {
	return dbutils.WithTx(ctx, r.db, func(tx *generated.Tx) (*model.PromotionCodeDTO, error) {
		if input == nil {
			return nil, errors.New("input is required")
		}

		exists, err := tx.PromotionCode.
			Query().
			Where(promotioncode.CodeEQ(input.Code)).
			Exist(ctx)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("promotion code already exists")
		}

		q := tx.PromotionCode.
			Create().
			SetCode(input.Code).
			SetDiscountType(input.DiscountType).
			SetDiscountValue(input.DiscountValue).
			SetIsActive(input.IsActive).
			SetStartAt(input.StartAt).
			SetEndAt(input.EndAt)

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

		promo, err := q.Save(ctx)
		if err != nil {
			return nil, err
		}

		// ===== Create Scopes =====
		for _, s := range input.Scopes {
			if err := promotionmodel.ValidateScopeInput(s.ScopeType); err != nil {
				_ = tx.Rollback()
				return nil, err
			}

			if _, err := tx.PromotionScope.
				Create().
				SetPromoCodeID(promo.ID).
				SetScopeType(s.ScopeType).
				SetScopeValue(s.ScopeValue).
				Save(ctx); err != nil {
				return nil, err
			}
		}

		// ===== Create Conditions =====
		for _, c := range input.Conditions {
			if err := promotionmodel.ValidateConditionInput(c.ConditionType); err != nil {
				return nil, err
			}

			if _, err := tx.PromotionCondition.
				Create().
				SetPromoCodeID(promo.ID).
				SetConditionType(c.ConditionType).
				SetConditionValue(c.ConditionValue).
				Save(ctx); err != nil {

				return nil, err
			}
		}

		return mapPromotionCodeDTOFromInput(promo, input.Scopes, input.Conditions), nil

	})
}

func (r *promotionRepository) UpdatePromotion(
	ctx context.Context,
	id int,
	input *model.UpdatePromotionInput,
) (*model.PromotionCodeDTO, error) {
	return dbutils.WithTx(ctx, r.db, func(tx *generated.Tx) (*model.PromotionCodeDTO, error) {
		if input == nil {
			return nil, errors.New("input is required")
		}
		if id <= 0 {
			return nil, errors.New("invalid id")
		}

		q := tx.PromotionCode.
			UpdateOneID(id).
			SetDiscountType(input.DiscountType).
			SetDiscountValue(input.DiscountValue)

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

		q.SetStartAt(input.StartAt)
		q.SetEndAt(input.EndAt)
		q.SetIsActive(input.IsActive)

		promo, err := q.Save(ctx)
		if err != nil {
			return nil, err
		}

		// ===== Replace Scopes =====
		if _, err := tx.PromotionScope.
			Delete().
			Where(promotionscope.PromoCodeID(id)).
			Exec(ctx); err != nil {
			return nil, err
		}

		for _, s := range input.Scopes {
			if err := promotionmodel.ValidateScopeInput(s.ScopeType); err != nil {
				return nil, err
			}

			if _, err := tx.PromotionScope.
				Create().
				SetPromoCodeID(id).
				SetScopeType(s.ScopeType).
				SetScopeValue(s.ScopeValue).
				Save(ctx); err != nil {
				return nil, err
			}
		}

		// ===== Replace Conditions =====
		if _, err := tx.PromotionCondition.
			Delete().
			Where(promotioncondition.PromoCodeID(id)).
			Exec(ctx); err != nil {
			return nil, err
		}

		for _, c := range input.Conditions {
			if err := promotionmodel.ValidateConditionInput(c.ConditionType); err != nil {
				return nil, err
			}

			if _, err := tx.PromotionCondition.
				Create().
				SetPromoCodeID(id).
				SetConditionType(c.ConditionType).
				SetConditionValue(c.ConditionValue).
				Save(ctx); err != nil {
				return nil, err
			}
		}

		return mapPromotionCodeDTOFromInput(promo, input.Scopes, input.Conditions), nil
	})
}

func (r *promotionRepository) DeletePromotion(ctx context.Context, id int) error {
	hasUsage, err := r.db.PromotionUsage.
		Query().
		Where(promotionusage.PromoCodeID(id)).
		Exist(ctx)
	if err != nil {
		return err
	}
	if hasUsage {
		return nil
	}

	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}

	if _, err = tx.PromotionScope.
		Delete().
		Where(promotionscope.PromoCodeID(id)).
		Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err = tx.PromotionCondition.
		Delete().
		Where(promotioncondition.PromoCodeID(id)).
		Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err = tx.PromotionCode.
		DeleteOneID(id).
		Exec(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}

func (r *promotionRepository) GetPromotionByID(
	ctx context.Context,
	id int,
) (*model.PromotionCodeDTO, error) {
	promo, err := r.db.PromotionCode.
		Query().
		Where(promotioncode.ID(id)).
		WithScopes().
		WithConditions().
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return mapPromotionCodeDTO(promo), nil
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
			return q.
				WithScopes().
				WithConditions()
		},
		func(src []*generated.PromotionCode) []*model.PromotionCodeDTO {
			out := make([]*model.PromotionCodeDTO, 0, len(src))
			for _, promo := range src {
				if promo == nil {
					continue
				}
				out = append(out, mapPromotionCodeDTO(promo))
			}
			return out
		},
	)
	if err != nil {
		var zero table.TableListResult[model.PromotionCodeDTO]
		return zero, err
	}
	return list, nil
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
