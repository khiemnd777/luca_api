package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/clinic"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/clinicdentist"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/dentist"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type ClinicRepository interface {
	Create(ctx context.Context, input model.ClinicDTO) (*model.ClinicDTO, error)
	Update(ctx context.Context, input model.ClinicDTO) (*model.ClinicDTO, error)
	GetByID(ctx context.Context, id int) (*model.ClinicDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.ClinicDTO], error)
	ListByDentistID(ctx context.Context, dentistID int, query table.TableQuery) (table.TableListResult[model.ClinicDTO], error)
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

	entity, err := tx.Clinic.Create().
		SetName(input.Name).
		SetNillableAddress(input.Address).
		SetNillablePhoneNumber(input.PhoneNumber).
		SetNillableBrief(input.Brief).
		SetNillableLogo(input.Logo).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	// Edge
	if input.DentistIDs != nil {
		dentistIDs := utils.DedupInt(input.DentistIDs, -1)
		if len(dentistIDs) > 0 {
			bulk := make([]*generated.ClinicDentistCreate, 0, len(dentistIDs))
			for _, did := range dentistIDs {
				bulk = append(bulk, tx.ClinicDentist.Create().
					SetClinicID(entity.ID).
					SetDentistID(did),
				)
			}
			if err = tx.ClinicDentist.CreateBulk(bulk...).Exec(ctx); err != nil {
				return nil, err
			}
		}
	}

	dto := mapper.MapAs[*generated.Clinic, *model.ClinicDTO](entity)
	return dto, nil
}

func (r *clinicRepo) Update(ctx context.Context, input model.ClinicDTO) (*model.ClinicDTO, error) {
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

	q := tx.Clinic.UpdateOneID(input.ID).
		SetName(input.Name).
		SetNillableAddress(input.Address).
		SetNillablePhoneNumber(input.PhoneNumber).
		SetNillableBrief(input.Brief).
		SetNillableLogo(input.Logo)

	// Edge
	if input.DentistIDs != nil {
		dentistIDs := utils.DedupInt(input.DentistIDs, -1)
		if _, err = tx.ClinicDentist.
			Delete().
			Where(clinicdentist.ClinicIDEQ(input.ID)).
			Exec(ctx); err != nil {
			return nil, err
		}
		if len(dentistIDs) > 0 {
			bulk := make([]*generated.ClinicDentistCreate, 0, len(dentistIDs))
			for _, did := range dentistIDs {
				bulk = append(bulk, tx.ClinicDentist.Create().
					SetClinicID(input.ID).
					SetDentistID(did),
				)
			}
			if err = tx.ClinicDentist.CreateBulk(bulk...).Exec(ctx); err != nil {
				return nil, err
			}
		}
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Clinic, *model.ClinicDTO](entity)
	return dto, nil
}

func (r *clinicRepo) GetByID(ctx context.Context, id int) (*model.ClinicDTO, error) {
	q := r.db.Clinic.Query().
		Where(
			clinic.ID(id),
			clinic.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)

	if err != nil {
		return nil, err
	}

	dentistIDs, err := q.
		QueryDentists().
		QueryDentist().
		Where(dentist.DeletedAtIsNil()).
		Order(dentist.ByID()).
		IDs(ctx)

	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Clinic, *model.ClinicDTO](entity)
	dto.DentistIDs = dentistIDs
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

func (r *clinicRepo) ListByDentistID(ctx context.Context, dentistID int, query table.TableQuery) (table.TableListResult[model.ClinicDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Clinic.Query().
			Where(
				clinic.HasDentistsWith(clinicdentist.DentistIDEQ(dentistID)),
				clinic.DeletedAtIsNil(),
			),
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
