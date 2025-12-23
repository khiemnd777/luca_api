package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/patient/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/module"
	searchmodel "github.com/khiemnd777/andy_api/shared/modules/search/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type PatientService interface {
	Create(ctx context.Context, deptID int, input model.PatientDTO) (*model.PatientDTO, error)
	Update(ctx context.Context, deptID int, input model.PatientDTO) (*model.PatientDTO, error)
	GetByID(ctx context.Context, id int) (*model.PatientDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.PatientDTO], error)
	ListByClinicID(ctx context.Context, clinicID int, query table.TableQuery) (table.TableListResult[model.PatientDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.PatientDTO], error)
	Delete(ctx context.Context, id int) error
}

type patientService struct {
	repo repository.PatientRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewPatientService(repo repository.PatientRepository, deps *module.ModuleDeps[config.ModuleConfig]) PatientService {
	return &patientService{repo: repo, deps: deps}
}

func kPatientByID(id int) string {
	return fmt.Sprintf("patient:id:%d", id)
}

func kPatientAll() []string {
	return []string{
		kPatientListAll(),
		kPatientSearchAll(),
		kPatientClinicAll(),
	}
}

func kPatientListAll() string {
	return "patient:list:*"
}

func kPatientSearchAll() string {
	return "patient:search:*"
}

func kPatientClinicAll() string {
	return "patient:clinic:*"
}

func kPatientClinicList(patientID int) string {
	return fmt.Sprintf("clinic:patient:%d:*", patientID)
}

func kPatientList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("patient:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func kClinicPatientList(clinicID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("patient:clinic:%d:list:l%d:p%d:o%s:d%s", clinicID, q.Limit, q.Page, orderBy, q.Direction)
}

func kPatientSearch(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("patient:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

func (s *patientService) Create(ctx context.Context, deptID int, input model.PatientDTO) (*model.PatientDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kPatientByID(dto.ID), kPatientClinicList(dto.ID))
	}

	cache.InvalidateKeys(kPatientAll()...)

	upsertSearch(deptID, dto)

	return dto, nil
}

func (s *patientService) Update(ctx context.Context, deptID int, input model.PatientDTO) (*model.PatientDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kPatientByID(dto.ID), kPatientClinicList(dto.ID))
	}
	cache.InvalidateKeys(kPatientAll()...)

	upsertSearch(deptID, dto)

	return dto, nil
}

func upsertSearch(deptID int, dto *model.PatientDTO) {
	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "patient",
		EntityID:   int64(dto.ID),
		Title:      dto.Name,
		Subtitle:   nil,
		Keywords:   utils.Ptr(*dto.PhoneNumber),
		Content:    nil,
		Attributes: nil,
		OrgID:      utils.Ptr(int64(deptID)),
		OwnerID:    utils.Ptr(int64(dto.ID)),
	})
}

func (s *patientService) GetByID(ctx context.Context, id int) (*model.PatientDTO, error) {
	return cache.Get(kPatientByID(id), cache.TTLMedium, func() (*model.PatientDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

func (s *patientService) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.PatientDTO], error) {
	type boxed = table.TableListResult[model.PatientDTO]
	key := kPatientList(q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.List(ctx, q)
		if e != nil {
			return nil, e
		}
		return &res, nil
	})
	if err != nil {
		var zero boxed
		return zero, err
	}
	return *ptr, nil
}

func (s *patientService) ListByClinicID(ctx context.Context, clinicID int, query table.TableQuery) (table.TableListResult[model.PatientDTO], error) {
	type boxed = table.TableListResult[model.PatientDTO]
	key := kClinicPatientList(clinicID, query)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.ListByClinicID(ctx, clinicID, query)
		if e != nil {
			return nil, e
		}
		return &res, nil
	})
	if err != nil {
		var zero boxed
		return zero, err
	}
	return *ptr, nil
}

func (s *patientService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kPatientAll()...)
	cache.InvalidateKeys(kPatientByID(id), kPatientClinicList(id))
	return nil
}

func (s *patientService) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.PatientDTO], error) {
	type boxed = dbutils.SearchResult[model.PatientDTO]
	key := kPatientSearch(q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.Search(ctx, q)
		if e != nil {
			return nil, e
		}
		return &res, nil
	})
	if err != nil {
		var zero boxed
		return zero, err
	}
	return *ptr, nil
}
