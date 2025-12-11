package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/product"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type ProductRepository interface {
	Create(ctx context.Context, input *model.ProductUpsertDTO) (*model.ProductDTO, error)
	Update(ctx context.Context, input *model.ProductUpsertDTO) (*model.ProductDTO, error)
	GetByID(ctx context.Context, id int) (*model.ProductDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.ProductDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.ProductDTO], error)
	Delete(ctx context.Context, id int) error
}

type productRepo struct {
	db    *generated.Client
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewProductRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) ProductRepository {
	return &productRepo{db: db, deps: deps, cfMgr: cfMgr}
}

func (r *productRepo) Create(ctx context.Context, input *model.ProductUpsertDTO) (*model.ProductDTO, error) {
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

	in := &input.DTO

	q := tx.Product.Create().
		SetNillableCode(in.Code).
		SetNillableName(in.Name).
		SetNillableCategoryID(in.CategoryID).
		SetNillableCategoryName(in.CategoryName)

	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err = customfields.PrepareCustomFields(ctx,
			r.cfMgr,
			*input.Collections,
			in.CustomFields,
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

	out := mapper.MapAs[*generated.Product, *model.ProductDTO](entity)

	if _, err = relation.UpsertM2M(ctx, tx, "products_processes", entity, input.DTO, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *productRepo) Update(ctx context.Context, input *model.ProductUpsertDTO) (*model.ProductDTO, error) {
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

	in := &input.DTO

	q := tx.Product.UpdateOneID(in.ID).
		SetNillableCode(in.Code).
		SetNillableName(in.Name).
		SetNillableCategoryID(in.CategoryID).
		SetNillableCategoryName(in.CategoryName)

	// custom fields
	if input.Collections != nil && len(*input.Collections) > 0 {
		_, err = customfields.PrepareCustomFields(ctx,
			r.cfMgr,
			*input.Collections,
			in.CustomFields,
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

	out := mapper.MapAs[*generated.Product, *model.ProductDTO](entity)

	if _, err = relation.UpsertM2M(ctx, tx, "products_processes", entity, input.DTO, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *productRepo) GetByID(ctx context.Context, id int) (*model.ProductDTO, error) {
	q := r.db.Product.Query().
		Where(
			product.ID(id),
			product.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Product, *model.ProductDTO](entity)
	return dto, nil
}

func (r *productRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.ProductDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Product.Query().
			Where(product.DeletedAtIsNil()),
		query,
		product.Table,
		product.FieldID,
		product.FieldID,
		func(src []*generated.Product) []*model.ProductDTO {
			return mapper.MapListAs[*generated.Product, *model.ProductDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.ProductDTO]
		return zero, err
	}
	return list, nil
}

func (r *productRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.ProductDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Product.Query().
			Where(product.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(product.FieldCode),
			dbutils.GetNormField(product.FieldName),
		},
		query,
		product.Table,
		product.FieldID,
		product.FieldID,
		product.Or,
		func(src []*generated.Product) []*model.ProductDTO {
			return mapper.MapListAs[*generated.Product, *model.ProductDTO](src)
		},
	)
}

func (r *productRepo) Delete(ctx context.Context, id int) error {
	return r.db.Product.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
