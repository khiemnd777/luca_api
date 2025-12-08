package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/materialsupplier"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/supplier"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type SupplierRepository interface {
	Create(ctx context.Context, input model.SupplierDTO) (*model.SupplierDTO, error)
	Update(ctx context.Context, input model.SupplierDTO) (*model.SupplierDTO, error)
	GetByID(ctx context.Context, id int) (*model.SupplierDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.SupplierDTO], error)
	ListByMaterialID(ctx context.Context, materialID int, query table.TableQuery) (table.TableListResult[model.SupplierDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.SupplierDTO], error)
	Delete(ctx context.Context, id int) error
}

type supplierRepo struct {
	db    *generated.Client
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewSupplierRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) SupplierRepository {
	return &supplierRepo{db: db, deps: deps, cfMgr: cfMgr}
}

func (r *supplierRepo) Create(ctx context.Context, input model.SupplierDTO) (*model.SupplierDTO, error) {
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

	q := tx.Supplier.Create().
		SetNillableCode(input.Code).
		SetNillableName(input.Name)

	_, err = customfields.PrepareCustomFields(ctx,
		r.cfMgr,
		[]string{"supplier"},
		input.CustomFields,
		q,
		false,
	)
	if err != nil {
		return nil, err
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Supplier, *model.SupplierDTO](entity)

	return dto, nil
}

func (r *supplierRepo) Update(ctx context.Context, input model.SupplierDTO) (*model.SupplierDTO, error) {
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

	q := tx.Supplier.UpdateOneID(input.ID).
		SetNillableCode(input.Code).
		SetNillableName(input.Name)

	_, err = customfields.PrepareCustomFields(ctx,
		r.cfMgr,
		[]string{"supplier"},
		input.CustomFields,
		q,
		false,
	)
	if err != nil {
		return nil, err
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Supplier, *model.SupplierDTO](entity)

	return dto, nil
}

func (r *supplierRepo) GetByID(ctx context.Context, id int) (*model.SupplierDTO, error) {
	q := r.db.Supplier.Query().
		Where(
			supplier.ID(id),
			supplier.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Supplier, *model.SupplierDTO](entity)
	return dto, nil
}

func (r *supplierRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.SupplierDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Supplier.Query().
			Where(supplier.DeletedAtIsNil()),
		query,
		supplier.Table,
		supplier.FieldID,
		supplier.FieldID,
		func(src []*generated.Supplier) []*model.SupplierDTO {
			return mapper.MapListAs[*generated.Supplier, *model.SupplierDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.SupplierDTO]
		return zero, err
	}
	return list, nil
}

func (r *supplierRepo) ListByMaterialID(ctx context.Context, materialID int, query table.TableQuery) (table.TableListResult[model.SupplierDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Supplier.Query().
			Where(supplier.HasMaterialsWith(materialsupplier.MaterialIDEQ(materialID)), supplier.DeletedAtIsNil()),
		query,
		supplier.Table,
		supplier.FieldID,
		supplier.FieldID,
		func(src []*generated.Supplier) []*model.SupplierDTO {
			mapped := mapper.MapListAs[*generated.Supplier, *model.SupplierDTO](src)
			return mapped
		},
	)
	if err != nil {
		var zero table.TableListResult[model.SupplierDTO]
		return zero, err
	}
	return list, nil
}

func (r *supplierRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.SupplierDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Supplier.Query().
			Where(supplier.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(supplier.FieldCode),
			dbutils.GetNormField(supplier.FieldName),
		},
		query,
		supplier.Table,
		supplier.FieldID,
		supplier.FieldID,
		supplier.Or,
		func(src []*generated.Supplier) []*model.SupplierDTO {
			return mapper.MapListAs[*generated.Supplier, *model.SupplierDTO](src)
		},
	)
}

func (r *supplierRepo) Delete(ctx context.Context, id int) error {
	return r.db.Supplier.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
