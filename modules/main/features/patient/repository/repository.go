package repository

import (
	"context"
	"time"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/clinic"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/clinicpatient"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/patient"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/mapper"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type PatientRepository interface {
	Create(ctx context.Context, input model.PatientDTO) (*model.PatientDTO, error)
	Update(ctx context.Context, input model.PatientDTO) (*model.PatientDTO, error)
	GetByID(ctx context.Context, id int) (*model.PatientDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.PatientDTO], error)
	ListByClinicID(ctx context.Context, clinicID int, query table.TableQuery) (table.TableListResult[model.PatientDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.PatientDTO], error)
	Delete(ctx context.Context, id int) error
}

type patientRepo struct {
	db   *generated.Client
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewPatientRepository(db *generated.Client, deps *module.ModuleDeps[config.ModuleConfig]) PatientRepository {
	return &patientRepo{db: db, deps: deps}
}

func (r *patientRepo) Create(ctx context.Context, input model.PatientDTO) (*model.PatientDTO, error) {
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

	entity, err := tx.Patient.Create().
		SetName(input.Name).
		SetNillablePhoneNumber(input.PhoneNumber).
		SetNillableBrief(input.Brief).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	// Edge
	if input.ClinicIDs != nil {
		clinicIDs := utils.DedupInt(input.ClinicIDs, -1)
		if len(clinicIDs) > 0 {
			bulk := make([]*generated.ClinicPatientCreate, 0, len(clinicIDs))
			for _, cid := range clinicIDs {
				bulk = append(bulk, tx.ClinicPatient.Create().
					SetPatientID(entity.ID).
					SetClinicID(cid),
				)
			}
			if err = tx.ClinicPatient.CreateBulk(bulk...).Exec(ctx); err != nil {
				return nil, err
			}
		}
	}

	dto := mapper.MapAs[*generated.Patient, *model.PatientDTO](entity)
	return dto, nil
}

func (r *patientRepo) Update(ctx context.Context, input model.PatientDTO) (*model.PatientDTO, error) {
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

	q := tx.Patient.UpdateOneID(input.ID).
		SetName(input.Name).
		SetNillablePhoneNumber(input.PhoneNumber).
		SetNillableBrief(input.Brief)

	// Edge
	if input.ClinicIDs != nil {
		clinicIDs := utils.DedupInt(input.ClinicIDs, -1)
		if _, err = tx.ClinicPatient.
			Delete().
			Where(clinicpatient.PatientIDEQ(input.ID)).
			Exec(ctx); err != nil {
			return nil, err
		}
		if len(clinicIDs) > 0 {
			bulk := make([]*generated.ClinicPatientCreate, 0, len(clinicIDs))
			for _, cid := range clinicIDs {
				bulk = append(bulk, tx.ClinicPatient.Create().
					SetPatientID(input.ID).
					SetClinicID(cid),
				)
			}
			if err = tx.ClinicPatient.CreateBulk(bulk...).Exec(ctx); err != nil {
				return nil, err
			}
		}
	}

	entity, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}

	dto := mapper.MapAs[*generated.Patient, *model.PatientDTO](entity)
	return dto, nil
}

func (r *patientRepo) GetByID(ctx context.Context, id int) (*model.PatientDTO, error) {
	q := r.db.Patient.Query().
		Where(
			patient.ID(id),
			patient.DeletedAtIsNil(),
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

	dto := mapper.MapAs[*generated.Patient, *model.PatientDTO](entity)
	dto.ClinicIDs = clinicIDs
	return dto, nil
}

func (r *patientRepo) List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.PatientDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Patient.Query().
			Where(patient.DeletedAtIsNil()),
		query,
		patient.Table,
		patient.FieldID,
		patient.FieldID,
		func(src []*generated.Patient) []*model.PatientDTO {
			mapped := mapper.MapListAs[*generated.Patient, *model.PatientDTO](src)
			return mapped
		},
	)
	if err != nil {
		var zero table.TableListResult[model.PatientDTO]
		return zero, err
	}
	return list, nil
}

func (r *patientRepo) ListByClinicID(ctx context.Context, clinicID int, query table.TableQuery) (table.TableListResult[model.PatientDTO], error) {
	list, err := table.TableList(
		ctx,
		r.db.Patient.Query().
			Where(patient.HasClinicsWith(clinicpatient.ClinicIDEQ(clinicID)), patient.DeletedAtIsNil()),
		query,
		patient.Table,
		patient.FieldID,
		patient.FieldID,
		func(src []*generated.Patient) []*model.PatientDTO {
			mapped := mapper.MapListAs[*generated.Patient, *model.PatientDTO](src)
			return mapped
		},
	)
	if err != nil {
		var zero table.TableListResult[model.PatientDTO]
		return zero, err
	}
	return list, nil
}

func (r *patientRepo) Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.PatientDTO], error) {
	return dbutils.Search(
		ctx,
		r.db.Patient.Query().
			Where(patient.DeletedAtIsNil()),
		[]string{
			dbutils.GetNormField(patient.FieldName),
		},
		query,
		patient.Table,
		patient.FieldID,
		patient.FieldID,
		patient.Or,
		func(src []*generated.Patient) []*model.PatientDTO {
			mapped := mapper.MapListAs[*generated.Patient, *model.PatientDTO](src)
			return mapped
		},
	)
}

func (r *patientRepo) Delete(ctx context.Context, id int) error {
	return r.db.Patient.UpdateOneID(id).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}
