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

type DentistRepository interface {
	Create(ctx context.Context, input model.DentistDTO) (*model.DentistDTO, error)
	Update(ctx context.Context, input model.DentistDTO) (*model.DentistDTO, error)
	GetByID(ctx context.Context, id int) (*model.DentistDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.DentistDTO], error)
	ListByClinicID(ctx context.Context, clinicID int, query table.TableQuery) (table.TableListResult[model.DentistDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.DentistDTO], error)
	Delete(ctx context.Context, id int) error
}

type dentistRepo struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewDentistRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) DentistRepository {
	return &dentistRepo{db: db, deps: deps}
}

func (r *dentistRepo) Create(ctx context.Context, input model.DentistDTO) (*model.DentistDTO, error) {
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

	q := tx.Dentist.Create().
		SetName(input.Name).
		SetNillablePhoneNumber(input.PhoneNumber).
		SetNillableBrief(input.Brief)

	// Edge
	clinicIDs := utils.DedupInt(input.ClinicIDs, -1)

	if len(clinicIDs) > 0 {
		q = q.AddClinicIDs(clinicIDs...)
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Dentist, *model.DentistDTO](entity)
	return dto, nil
}

func (r *dentistRepo) Update(ctx context.Context, input model.DentistDTO) (*model.DentistDTO, error) {
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

	q := tx.Dentist.UpdateOneID(input.ID).
		SetName(input.Name).
		SetNillablePhoneNumber(input.PhoneNumber).
		SetNillableBrief(input.Brief)

	// Edge
	if input.ClinicIDs != nil {
		clinicIDs := utils.DedupInt(input.ClinicIDs, -1)
		if _, err = tx.ClinicDentist.
			Delete().
			Where(clinicdentist.DentistIDEQ(input.ID)).
			Exec(ctx); err != nil {
			return nil, err
		}
		if len(clinicIDs) > 0 {
			bulk := make([]*generated.ClinicDentistCreate, 0, len(clinicIDs))
			for _, cid := range clinicIDs {
				bulk = append(bulk, tx.ClinicDentist.Create().
					SetDentistID(input.ID).
					SetClinicID(cid),
				)
				if err = tx.ClinicDentist.CreateBulk(bulk...).Exec(ctx); err != nil {
					return nil, err
				}
			}
		}
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Dentist, *model.DentistDTO](entity)
	return dto, nil
}

func (r *dentistRepo) GetByID(ctx context.Context, id int) (*model.DentistDTO, error) {
	q := r.db.Dentist.Query().
		Where(
			dentist.ID(id),
			dentist.DeletedAtIsNil(),
		)

	entity, err := q.Only(ctx)
	if err != nil {
		return nil, err
	}

	clinicIDs, err := q.
		QueryClinics().
		QueryClinic().
		Where(clinic.DeletedAtIsNil()).
		Order(clinic.ByID()).
		IDs(ctx)

	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Dentist, *model.DentistDTO](entity)
	dto.ClinicIDs = clinicIDs
	return dto, nil
}

func (r *dentistRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.DentistDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Dentist.Query().
			Where(dentist.DeletedAtIsNil()),
		query,
		dentist.Table,
		dentist.FieldID,
		dentist.FieldID,
		func(src []*generated.Dentist) []*model.DentistDTO {
			mapped := mapper.MapListAs[*generated.Dentist, *model.DentistDTO](src)
			return mapped
		},
	)
	if err != nil {
		var zero table.TableListResult[model.DentistDTO]
		return zero, err
	}
	return list, nil
}

func (r *dentistRepo) ListByClinicID(ctx context.Context, clinicID int, query table.TableQuery) (table.TableListResult[model.DentistDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Dentist.Query().
			Where(dentist.HasClinicsWith(clinicdentist.ClinicIDEQ(clinicID)), dentist.DeletedAtIsNil()),
		query,
		dentist.Table,
		dentist.FieldID,
		dentist.FieldID,
		func(src []*generated.Dentist) []*model.DentistDTO {
			mapped := mapper.MapListAs[*generated.Dentist, *model.DentistDTO](src)
			return mapped
		},
	)
	if err != nil {
		var zero table.TableListResult[model.DentistDTO]
		return zero, err
	}
	return list, nil
}

func (r *dentistRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.DentistDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Dentist.Query().
			Where(dentist.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(dentist.FieldName),
		},
		query,
		dentist.Table,
		dentist.FieldID,
		dentist.FieldID,
		dentist.Or,
		func(src []*generated.Dentist) []*model.DentistDTO {
			mapped := mapper.MapListAs[*generated.Dentist, *model.DentistDTO](src)
			return mapped
		},
	)
}

func (r *dentistRepo) Delete(ctx context.Context, id int) error {
	return r.db.Dentist.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
