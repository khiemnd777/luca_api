package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	"github.com/khiemnd777/andy_api/modules/main/features/clinic/model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/clinic"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type ClinicRepository interface {
	Create(ctx context.Context, input model.ClinicDTO) (*model.ClinicDTO, error)
	Update(ctx context.Context, input model.ClinicDTO) (*model.ClinicDTO, error)
	GetByID(ctx context.Context, id int) (*model.ClinicDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.ClinicDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.ClinicDTO], error)
	Delete(ctx context.Context, id int) error
}

type clinicRepo struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewClinicRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) ClinicRepository {
	return &clinicRepo{db: db, deps: deps}
}

func (r *clinicRepo) Create(ctx context.Context, input model.ClinicDTO) (*model.ClinicDTO, error) {
	q := r.db.Clinic.Create().
		SetActive(input.Active).
		SetName(input.Name).
		SetNillableBrief(input.Brief).
		SetNillableLogo(input.Logo)

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Clinic, *model.ClinicDTO](entity)
	return dto, nil
}

func (r *clinicRepo) Update(ctx context.Context, input model.ClinicDTO) (*model.ClinicDTO, error) {
	entity, err := r.db.Clinic.UpdateOneID(input.ID).
		SetActive(input.Active).
		SetName(input.Name).
		SetNillableBrief(input.Brief).
		SetNillableLogo(input.Logo).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Clinic, *model.ClinicDTO](entity)
	return dto, nil
}

func (r *clinicRepo) GetByID(ctx context.Context, id int) (*model.ClinicDTO, error) {
	entity, err := r.db.Clinic.Query().
		Where(
			clinic.ID(id),
			clinic.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Clinic, *model.ClinicDTO](entity)
	return dto, nil
}

func (r *clinicRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.ClinicDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Clinic.Query().
			Where(clinic.DeletedAtIsNil()),
		query,
		clinic.Table,
		clinic.FieldID,
		clinic.FieldID,
		func(src []*generated.Clinic) []*model.ClinicDTO {
			mapped := mapper.MapListAs[*generated.Clinic, *model.ClinicDTO](src)
			return mapped
		},
	)
	if err != nil {
		var zero table.TableListResult[model.ClinicDTO]
		return zero, err
	}
	return list, nil
}

func (r *clinicRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.ClinicDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Clinic.Query().
			Where(clinic.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(clinic.FieldName),
		},
		query,
		clinic.Table,
		clinic.FieldID,
		clinic.FieldID,
		clinic.Or,
		func(src []*generated.Clinic) []*model.ClinicDTO {
			mapped := mapper.MapListAs[*generated.Clinic, *model.ClinicDTO](src)
			return mapped
		},
	)
}

func (r *clinicRepo) Delete(ctx context.Context, id int) error {
	return r.db.Clinic.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
