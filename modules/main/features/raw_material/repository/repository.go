package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/category"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/rawmaterial"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type RawMaterialRepository interface {
	Create(ctx context.Context, deptID int, input model.RawMaterialDTO) (*model.RawMaterialDTO, error)
	Update(ctx context.Context, deptID int, input model.RawMaterialDTO) (*model.RawMaterialDTO, error)
	GetByID(ctx context.Context, deptID int, id int) (*model.RawMaterialDTO, error)
	List(ctx context.Context, deptID int, categoryID *int, query table.TableQuery) (table.TableListResult[model.RawMaterialDTO], error)
	Search(ctx context.Context, deptID int, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.RawMaterialDTO], error)
	Delete(ctx context.Context, deptID int, id int) error
}

type rawMaterialRepo struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewRawMaterialRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) RawMaterialRepository {
	return &rawMaterialRepo{db: db, deps: deps}
}

func (r *rawMaterialRepo) Create(ctx context.Context, deptID int, input model.RawMaterialDTO) (*model.RawMaterialDTO, error) {
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

	categoryName := input.CategoryName
	if categoryName == nil && input.CategoryID != nil {
		cat, err := tx.Category.Query().
			Where(
				category.ID(*input.CategoryID),
				category.DeletedAtIsNil(),
			).
			Only(ctx)
		if err != nil {
			return nil, err
		}
		categoryName = cat.Name
	}

	entity, err := tx.RawMaterial.Create().
		SetNillableDepartmentID(&deptID).
		SetNillableCategoryID(input.CategoryID).
		SetNillableCategoryName(categoryName).
		SetNillableName(input.Name).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.RawMaterial, *model.RawMaterialDTO](entity)
	return dto, nil
}

func (r *rawMaterialRepo) Update(ctx context.Context, deptID int, input model.RawMaterialDTO) (*model.RawMaterialDTO, error) {
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

	categoryName := input.CategoryName
	if categoryName == nil && input.CategoryID != nil {
		cat, err := tx.Category.Query().
			Where(
				category.ID(*input.CategoryID),
				category.DeletedAtIsNil(),
			).
			Only(ctx)
		if err != nil {
			return nil, err
		}
		categoryName = cat.Name
	}

	entity, err := tx.RawMaterial.UpdateOneID(input.ID).
		SetNillableDepartmentID(&deptID).
		SetNillableCategoryID(input.CategoryID).
		SetNillableCategoryName(categoryName).
		SetNillableName(input.Name).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.RawMaterial, *model.RawMaterialDTO](entity)
	return dto, nil
}

func (r *rawMaterialRepo) GetByID(ctx context.Context, deptID int, id int) (*model.RawMaterialDTO, error) {
	entity, err := r.db.RawMaterial.Query().
		Where(
			rawmaterial.ID(id),
			rawmaterial.DepartmentIDEQ(deptID),
			rawmaterial.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.RawMaterial, *model.RawMaterialDTO](entity)
	return dto, nil
}

func (r *rawMaterialRepo) List(ctx context.Context, deptID int, categoryID *int, query table.TableQuery) (table.TableListResult[model.RawMaterialDTO], error) {
	q := r.db.RawMaterial.Query().
		Where(
			rawmaterial.DeletedAtIsNil(),
			rawmaterial.DepartmentIDEQ(deptID),
		)
	if categoryID != nil {
		q = q.Where(rawmaterial.CategoryIDEQ(*categoryID))
	}

	list, err := table.TableList(
		ctx,
		q,
		query,
		rawmaterial.Table,
		rawmaterial.FieldID,
		rawmaterial.FieldID,
		func(src []*generated.RawMaterial) []*model.RawMaterialDTO {
			return mapper.MapListAs[*generated.RawMaterial, *model.RawMaterialDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.RawMaterialDTO]
		return zero, err
	}
	return list, nil
}

func (r *rawMaterialRepo) Search(ctx context.Context, deptID int, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.RawMaterialDTO], error) {
	q := r.db.RawMaterial.Query().
		Where(
			rawmaterial.DeletedAtIsNil(),
			rawmaterial.DepartmentIDEQ(deptID),
		)
	if categoryID != nil {
		q = q.Where(rawmaterial.CategoryIDEQ(*categoryID))
	}

	return dbutils.Search(
		ctx,
		q,
		[]string{
			dbutils.GetNormField(rawmaterial.FieldName),
		},
		query,
		rawmaterial.Table,
		rawmaterial.FieldID,
		rawmaterial.FieldID,
		rawmaterial.Or,
		func(src []*generated.RawMaterial) []*model.RawMaterialDTO {
			return mapper.MapListAs[*generated.RawMaterial, *model.RawMaterialDTO](src)
		},
	)
}

func (r *rawMaterialRepo) Delete(ctx context.Context, deptID int, id int) error {
	return r.db.RawMaterial.UpdateOneID(id).
		Where(rawmaterial.DepartmentIDEQ(deptID)).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
