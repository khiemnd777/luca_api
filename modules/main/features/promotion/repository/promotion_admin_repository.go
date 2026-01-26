package repository

import (
	"context"
	"database/sql"
	"errors"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	promotionmodel "github.com/khiemnd777/andy_api/modules/main/features/promotion/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotioncode"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotioncondition"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotionscope"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/promotionusage"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type PromotionAdminRepository interface {
	CreatePromotion(ctx context.Context, input *model.CreatePromotionInput) (*model.PromotionCodeDTO, error)
	UpdatePromotion(ctx context.Context, id int, input *model.UpdatePromotionInput) (*model.PromotionCodeDTO, error)
	DeletePromotion(ctx context.Context, id int) error
	GetPromotionByID(ctx context.Context, id int) (*model.PromotionCodeDTO, error)
	ListPromotions(ctx context.Context, query table.TableQuery) (table.TableListResult[model.PromotionCodeDTO], error)
}

type promotionAdminRepository struct {
	db    *generated.Client
	sqlDB *sql.DB
}

func NewPromotionAdminRepository(db *generated.Client, sqlDB *sql.DB) PromotionAdminRepository {
	return &promotionAdminRepository{db: db, sqlDB: sqlDB}
}

func (r *promotionAdminRepository) CreatePromotion(
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
			SetNillableEndAt(input.EndAt).
			SetNillableMaxDiscountAmount(input.MaxDiscountAmount).
			SetNillableMinOrderValue(input.MinOrderValue).
			SetNillableTotalUsageLimit(input.TotalUsageLimit).
			SetNillableUsagePerUser(input.UsagePerUser)

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

func (r *promotionAdminRepository) UpdatePromotion(
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
			SetDiscountValue(input.DiscountValue).
			SetNillableMaxDiscountAmount(input.MaxDiscountAmount).
			SetNillableMinOrderValue(input.MinOrderValue).
			SetNillableTotalUsageLimit(input.TotalUsageLimit).
			SetNillableUsagePerUser(input.UsagePerUser).
			SetStartAt(input.StartAt).
			SetNillableEndAt(input.EndAt).
			SetIsActive(input.IsActive)

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

func (r *promotionAdminRepository) DeletePromotion(ctx context.Context, id int) error {
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

func (r *promotionAdminRepository) GetPromotionByID(
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

func (r *promotionAdminRepository) ListPromotions(
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
