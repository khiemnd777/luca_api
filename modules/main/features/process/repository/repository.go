package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/process"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type ProcessRepository interface {
	Create(ctx context.Context, input model.ProcessDTO) (*model.ProcessDTO, error)
	Update(ctx context.Context, input model.ProcessDTO) (*model.ProcessDTO, error)
	GetByID(ctx context.Context, id int) (*model.ProcessDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.ProcessDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.ProcessDTO], error)
	Delete(ctx context.Context, id int) error
}

type processRepo struct {
	db    *generated.Client
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewProcessRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) ProcessRepository {
	return &processRepo{db: db, deps: deps, cfMgr: cfMgr}
}

func (r *processRepo) Create(ctx context.Context, input model.ProcessDTO) (*model.ProcessDTO, error) {
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

	q := tx.Process.Create().
		SetNillableCode(input.Code).
		SetNillableName(input.Name)

	_, err = customfields.PrepareCustomFields(ctx,
		r.cfMgr,
		[]string{"process"},
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

	dto := mapper.MapAs[*generated.Process, *model.ProcessDTO](entity)

	_, err = relation.Upsert(ctx, tx, "process", entity, input, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *processRepo) Update(ctx context.Context, input model.ProcessDTO) (*model.ProcessDTO, error) {
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

	q := tx.Process.UpdateOneID(input.ID).
		SetNillableCode(input.Code).
		SetNillableName(input.Name)

	_, err = customfields.PrepareCustomFields(ctx,
		r.cfMgr,
		[]string{"process"},
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

	dto := mapper.MapAs[*generated.Process, *model.ProcessDTO](entity)

	_, err = relation.Upsert(ctx, tx, "process", entity, input, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *processRepo) GetByID(ctx context.Context, id int) (*model.ProcessDTO, error) {
	q := r.db.Process.Query().
		Where(
			process.ID(id),
			process.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Process, *model.ProcessDTO](entity)
	return dto, nil
}

func (r *processRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.ProcessDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Process.Query().
			Where(process.DeletedAtIsNil()),
		query,
		process.Table,
		process.FieldID,
		process.FieldID,
		func(src []*generated.Process) []*model.ProcessDTO {
			return mapper.MapListAs[*generated.Process, *model.ProcessDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.ProcessDTO]
		return zero, err
	}
	return list, nil
}

func (r *processRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.ProcessDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Process.Query().
			Where(process.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(process.FieldCode),
			dbutils.GetNormField(process.FieldName),
		},
		query,
		process.Table,
		process.FieldID,
		process.FieldID,
		process.Or,
		func(src []*generated.Process) []*model.ProcessDTO {
			return mapper.MapListAs[*generated.Process, *model.ProcessDTO](src)
		},
	)
}

func (r *processRepo) Delete(ctx context.Context, id int) error {
	return r.db.Process.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
