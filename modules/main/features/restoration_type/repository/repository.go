package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/restorationtype"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type RestorationTypeRepository interface {
	Create(ctx context.Context, input model.RestorationTypeDTO) (*model.RestorationTypeDTO, error)
	Update(ctx context.Context, input model.RestorationTypeDTO) (*model.RestorationTypeDTO, error)
	GetByID(ctx context.Context, id int) (*model.RestorationTypeDTO, error)
	List(ctx context.Context, categoryID *int, query table.TableQuery) (table.TableListResult[model.RestorationTypeDTO], error)
	Search(ctx context.Context, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.RestorationTypeDTO], error)
	Delete(ctx context.Context, id int) error
}

type restorationTypeRepo struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewRestorationTypeRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) RestorationTypeRepository {
	return &restorationTypeRepo{db: db, deps: deps}
}

func (r *restorationTypeRepo) Create(ctx context.Context, input model.RestorationTypeDTO) (*model.RestorationTypeDTO, error) {
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

	entity, err := tx.RestorationType.Create().
		SetNillableCategoryID(input.CategoryID).
		SetNillableName(input.Name).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.RestorationType, *model.RestorationTypeDTO](entity)
	return dto, nil
}

func (r *restorationTypeRepo) Update(ctx context.Context, input model.RestorationTypeDTO) (*model.RestorationTypeDTO, error) {
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

	entity, err := tx.RestorationType.UpdateOneID(input.ID).
		SetNillableCategoryID(input.CategoryID).
		SetNillableName(input.Name).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.RestorationType, *model.RestorationTypeDTO](entity)
	return dto, nil
}

func (r *restorationTypeRepo) GetByID(ctx context.Context, id int) (*model.RestorationTypeDTO, error) {
	entity, err := r.db.RestorationType.Query().
		Where(
			restorationtype.ID(id),
			restorationtype.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.RestorationType, *model.RestorationTypeDTO](entity)
	return dto, nil
}

func (r *restorationTypeRepo) List(ctx context.Context, categoryID *int, query table.TableQuery) (table.TableListResult[model.RestorationTypeDTO], error) {
	q := r.db.RestorationType.Query().
		Where(restorationtype.DeletedAtIsNil())
	if categoryID != nil {
		q = q.Where(restorationtype.CategoryIDEQ(*categoryID))
	}

	list, err := table.TableList(
		ctx,
		q,
		query,
		restorationtype.Table,
		restorationtype.FieldID,
		restorationtype.FieldID,
		func(src []*generated.RestorationType) []*model.RestorationTypeDTO {
			return mapper.MapListAs[*generated.RestorationType, *model.RestorationTypeDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.RestorationTypeDTO]
		return zero, err
	}
	return list, nil
}

func (r *restorationTypeRepo) Search(ctx context.Context, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.RestorationTypeDTO], error) {
	q := r.db.RestorationType.Query().
		Where(restorationtype.DeletedAtIsNil())
	if categoryID != nil {
		q = q.Where(restorationtype.CategoryIDEQ(*categoryID))
	}

	return dbutils.Search(
		ctx,
		q,
		[]string{
			dbutils.GetNormField(restorationtype.FieldName),
		},
		query,
		restorationtype.Table,
		restorationtype.FieldID,
		restorationtype.FieldID,
		restorationtype.Or,
		func(src []*generated.RestorationType) []*model.RestorationTypeDTO {
			return mapper.MapListAs[*generated.RestorationType, *model.RestorationTypeDTO](src)
		},
	)
}

func (r *restorationTypeRepo) Delete(ctx context.Context, id int) error {
	return r.db.RestorationType.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
