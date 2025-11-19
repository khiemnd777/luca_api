package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/material"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type MaterialRepository interface {
	Create(ctx context.Context, input model.MaterialDTO) (*model.MaterialDTO, error)
	Update(ctx context.Context, input model.MaterialDTO) (*model.MaterialDTO, error)
	GetByID(ctx context.Context, id int) (*model.MaterialDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.MaterialDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.MaterialDTO], error)
	Delete(ctx context.Context, id int) error
}

type materialRepo struct {
	db    *generated.Client
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewMaterialRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) MaterialRepository {
	return &materialRepo{db: db, deps: deps, cfMgr: cfMgr}
}

func (r *materialRepo) Create(ctx context.Context, input model.MaterialDTO) (*model.MaterialDTO, error) {
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

	q := tx.Material.Create().
		SetNillableCode(input.Code).
		SetNillableName(input.Name)

	err = customfields.SetCustomFields(ctx, r.cfMgr, "material", input.CustomFields, q, false)
	if err != nil {
		return nil, err
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Material, *model.MaterialDTO](entity)

	_, err = relation.Upsert(ctx, tx, "material", entity, input, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *materialRepo) Update(ctx context.Context, input model.MaterialDTO) (*model.MaterialDTO, error) {
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

	q := tx.Material.UpdateOneID(input.ID).
		SetNillableCode(input.Code).
		SetNillableName(input.Name)

	err = customfields.SetCustomFields(ctx, r.cfMgr, "material", input.CustomFields, q, false)
	if err != nil {
		return nil, err
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Material, *model.MaterialDTO](entity)
	_, err = relation.Upsert(ctx, tx, "material", entity, input, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *materialRepo) GetByID(ctx context.Context, id int) (*model.MaterialDTO, error) {
	q := r.db.Material.Query().
		Where(
			material.ID(id),
			material.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Material, *model.MaterialDTO](entity)
	return dto, nil
}

func (r *materialRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.MaterialDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Material.Query().
			Where(material.DeletedAtIsNil()),
		query,
		material.Table,
		material.FieldID,
		material.FieldID,
		func(src []*generated.Material) []*model.MaterialDTO {
			return mapper.MapListAs[*generated.Material, *model.MaterialDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.MaterialDTO]
		return zero, err
	}
	return list, nil
}

func (r *materialRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.MaterialDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Material.Query().
			Where(material.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(material.FieldCode),
			dbutils.GetNormField(material.FieldName),
		},
		query,
		material.Table,
		material.FieldID,
		material.FieldID,
		material.Or,
		func(src []*generated.Material) []*model.MaterialDTO {
			return mapper.MapListAs[*generated.Material, *model.MaterialDTO](src)
		},
	)
}

func (r *materialRepo) Delete(ctx context.Context, id int) error {
	return r.db.Material.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
