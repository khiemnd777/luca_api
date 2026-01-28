package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/brandname"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type BrandNameRepository interface {
	Create(ctx context.Context, input model.BrandNameDTO) (*model.BrandNameDTO, error)
	Update(ctx context.Context, input model.BrandNameDTO) (*model.BrandNameDTO, error)
	GetByID(ctx context.Context, id int) (*model.BrandNameDTO, error)
	List(ctx context.Context, categoryID *int, query table.TableQuery) (table.TableListResult[model.BrandNameDTO], error)
	Search(ctx context.Context, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.BrandNameDTO], error)
	Delete(ctx context.Context, id int) error
}

type brandNameRepo struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewBrandNameRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) BrandNameRepository {
	return &brandNameRepo{db: db, deps: deps}
}

func (r *brandNameRepo) Create(ctx context.Context, input model.BrandNameDTO) (*model.BrandNameDTO, error) {
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

	entity, err := tx.BrandName.Create().
		SetNillableCategoryID(input.CategoryID).
		SetNillableName(input.Name).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.BrandName, *model.BrandNameDTO](entity)
	return dto, nil
}

func (r *brandNameRepo) Update(ctx context.Context, input model.BrandNameDTO) (*model.BrandNameDTO, error) {
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

	entity, err := tx.BrandName.UpdateOneID(input.ID).
		SetNillableCategoryID(input.CategoryID).
		SetNillableName(input.Name).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.BrandName, *model.BrandNameDTO](entity)
	return dto, nil
}

func (r *brandNameRepo) GetByID(ctx context.Context, id int) (*model.BrandNameDTO, error) {
	entity, err := r.db.BrandName.Query().
		Where(
			brandname.ID(id),
			brandname.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.BrandName, *model.BrandNameDTO](entity)
	return dto, nil
}

func (r *brandNameRepo) List(ctx context.Context, categoryID *int, query table.TableQuery) (table.TableListResult[model.BrandNameDTO], error) {
	q := r.db.BrandName.Query().
		Where(brandname.DeletedAtIsNil())
	if categoryID != nil {
		q = q.Where(brandname.CategoryIDEQ(*categoryID))
	}

	list, err := table.TableList(
		ctx,
		q,
		query,
		brandname.Table,
		brandname.FieldID,
		brandname.FieldID,
		func(src []*generated.BrandName) []*model.BrandNameDTO {
			return mapper.MapListAs[*generated.BrandName, *model.BrandNameDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.BrandNameDTO]
		return zero, err
	}
	return list, nil
}

func (r *brandNameRepo) Search(ctx context.Context, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.BrandNameDTO], error) {
	q := r.db.BrandName.Query().
		Where(brandname.DeletedAtIsNil())
	if categoryID != nil {
		q = q.Where(brandname.CategoryIDEQ(*categoryID))
	}

	return dbutils.Search(
		ctx,
		q,
		[]string{
			dbutils.GetNormField(brandname.FieldName),
		},
		query,
		brandname.Table,
		brandname.FieldID,
		brandname.FieldID,
		brandname.Or,
		func(src []*generated.BrandName) []*model.BrandNameDTO {
			return mapper.MapListAs[*generated.BrandName, *model.BrandNameDTO](src)
		},
	)
}

func (r *brandNameRepo) Delete(ctx context.Context, id int) error {
	return r.db.BrandName.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
