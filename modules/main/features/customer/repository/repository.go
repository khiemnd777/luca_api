package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/customer"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type CustomerRepository interface {
	Create(ctx context.Context, input model.CustomerDTO) (*model.CustomerDTO, error)
	Update(ctx context.Context, input model.CustomerDTO) (*model.CustomerDTO, error)
	GetByID(ctx context.Context, id int) (*model.CustomerDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.CustomerDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.CustomerDTO], error)
	Delete(ctx context.Context, id int) error
}

type customerRepo struct {
	db    *generated.Client
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewCustomerRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) CustomerRepository {
	return &customerRepo{db: db, deps: deps, cfMgr: cfMgr}
}

func (r *customerRepo) Create(ctx context.Context, input model.CustomerDTO) (*model.CustomerDTO, error) {
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

	q := tx.Customer.Create().
		SetNillableCode(input.Code).
		SetNillableName(input.Name)

	_, err = customfields.PrepareCustomFields(ctx,
		r.cfMgr,
		[]string{"customer"},
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

	dto := mapper.MapAs[*generated.Customer, *model.CustomerDTO](entity)

	_, err = relation.UpsertM2M(ctx, tx, "customer", entity, input, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *customerRepo) Update(ctx context.Context, input model.CustomerDTO) (*model.CustomerDTO, error) {
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

	q := tx.Customer.UpdateOneID(input.ID).
		SetNillableCode(input.Code).
		SetNillableName(input.Name)

	_, err = customfields.PrepareCustomFields(ctx,
		r.cfMgr,
		[]string{"customer"},
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

	dto := mapper.MapAs[*generated.Customer, *model.CustomerDTO](entity)

	_, err = relation.UpsertM2M(ctx, tx, "customer", entity, input, dto)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (r *customerRepo) GetByID(ctx context.Context, id int) (*model.CustomerDTO, error) {
	q := r.db.Customer.Query().
		Where(
			customer.ID(id),
			customer.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Customer, *model.CustomerDTO](entity)
	return dto, nil
}

func (r *customerRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.CustomerDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Customer.Query().
			Where(customer.DeletedAtIsNil()),
		query,
		customer.Table,
		customer.FieldID,
		customer.FieldID,
		func(src []*generated.Customer) []*model.CustomerDTO {
			return mapper.MapListAs[*generated.Customer, *model.CustomerDTO](src)
		},
	)
	if err != nil {
		var zero table.TableListResult[model.CustomerDTO]
		return zero, err
	}
	return list, nil
}

func (r *customerRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.CustomerDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Customer.Query().
			Where(customer.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(customer.FieldCode),
			dbutils.GetNormField(customer.FieldName),
		},
		query,
		customer.Table,
		customer.FieldID,
		customer.FieldID,
		customer.Or,
		func(src []*generated.Customer) []*model.CustomerDTO {
			return mapper.MapListAs[*generated.Customer, *model.CustomerDTO](src)
		},
	)
}

func (r *customerRepo) Delete(ctx context.Context, id int) error {
	return r.db.Customer.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
