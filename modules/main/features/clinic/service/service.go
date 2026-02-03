package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/clinic/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/metadata/customfields"
	"github.com/khiemnd777/andy_api/shared/module"
	searchmodel "github.com/khiemnd777/andy_api/shared/modules/search/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	searchutils "github.com/khiemnd777/andy_api/shared/search"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type ClinicService interface {
	Create(ctx context.Context, deptID int, input model.ClinicDTO) (*model.ClinicDTO, error)
	Update(ctx context.Context, deptID int, input model.ClinicDTO) (*model.ClinicDTO, error)
	GetByID(ctx context.Context, deptID int, id int) (*model.ClinicDTO, error)
	List(ctx context.Context, deptID int, query table.TableQuery) (table.TableListResult[model.ClinicDTO], error)
	ListByDentistID(ctx context.Context, deptID int, dentistID int, query table.TableQuery) (table.TableListResult[model.ClinicDTO], error)
	ListByPatientID(ctx context.Context, deptID int, patientID int, query table.TableQuery) (table.TableListResult[model.ClinicDTO], error)
	Search(ctx context.Context, deptID int, query dbutils.SearchQuery) (dbutils.SearchResult[model.ClinicDTO], error)
	Delete(ctx context.Context, deptID int, id int) error
}

type clinicService struct {
	repo  repository.ClinicRepository
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewClinicService(repo repository.ClinicRepository, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) ClinicService {
	return &clinicService{repo: repo, deps: deps, cfMgr: cfMgr}
}

func kClinicByID(deptID int, id int) string {
	return fmt.Sprintf("clinic:dpt%d:id:%d", deptID, id)
}

func kClinicAll(deptID int) []string {
	return []string{
		kClinicListAll(deptID),
		kClinicSearchAll(deptID),
		kClinicDentistAll(),
		kClinicPatientAll(),
	}
}

func kClinicListAll(deptID int) string {
	return fmt.Sprintf("clinic:list:dpt%d:*", deptID)
}

func kClinicSearchAll(deptID int) string {
	return fmt.Sprintf("clinic:search:dpt%d:*", deptID)
}

func kClinicDentistAll() string {
	return "clinic:dentist:*"
}

func kClinicPatientAll() string {
	return "clinic:patient:*"
}

func kClinicDentistList(clinicID int) string {
	return fmt.Sprintf("dentist:clinic:%d:*", clinicID)
}

func kClinicPatientList(clinicID int) string {
	return fmt.Sprintf("patient:clinic:%d:*", clinicID)
}

func kClinicList(deptID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("clinic:list:dpt%d:l%d:p%d:o%s:d%s", deptID, q.Limit, q.Page, orderBy, q.Direction)
}

func kDentistClinicList(deptID int, dentistID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("clinic:dpt%d:dentist:%d:list:l%d:p%d:o%s:d%s", deptID, dentistID, q.Limit, q.Page, orderBy, q.Direction)
}

func kPatientClinicList(deptID int, dentistID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("clinic:dpt%d:patient:%d:list:l%d:p%d:o%s:d%s", deptID, dentistID, q.Limit, q.Page, orderBy, q.Direction)
}

func kClinicSearch(deptID int, q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("clinic:search:dpt%d:k%s:l%d:p%d:o%s:d%s", deptID, q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

func (s *clinicService) Create(ctx context.Context, deptID int, input model.ClinicDTO) (*model.ClinicDTO, error) {
	dto, err := s.repo.Create(ctx, deptID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kClinicByID(deptID, dto.ID), kClinicDentistList(dto.ID), kClinicPatientList(dto.ID))
	}
	cache.InvalidateKeys(kClinicAll(deptID)...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

func (s *clinicService) Update(ctx context.Context, deptID int, input model.ClinicDTO) (*model.ClinicDTO, error) {
	dto, err := s.repo.Update(ctx, deptID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kClinicByID(deptID, dto.ID), kClinicDentistList(dto.ID), kClinicPatientList(dto.ID))
	}
	cache.InvalidateKeys(kClinicAll(deptID)...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

func (s *clinicService) upsertSearch(ctx context.Context, deptID int, dto *model.ClinicDTO) {
	kwPtr, _ := searchutils.BuildKeywords(ctx, s.cfMgr, "clinic", []any{dto.PhoneNumber}, dto.CustomFields)

	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "clinic",
		EntityID:   int64(dto.ID),
		Title:      dto.Name,
		Subtitle:   nil,
		Keywords:   &kwPtr,
		Content:    utils.Ptr(*dto.Brief),
		Attributes: map[string]any{
			"logo": dto.Logo,
		},
		OrgID:   utils.Ptr(int64(deptID)),
		OwnerID: utils.Ptr(int64(dto.ID)),
	})
}

func (s *clinicService) unlinkSearch(id int) {
	pubsub.PublishAsync("search:unlink", &searchmodel.UnlinkDoc{
		EntityType: "clinic",
		EntityID:   int64(id),
	})
}

func (s *clinicService) GetByID(ctx context.Context, deptID int, id int) (*model.ClinicDTO, error) {
	return cache.Get(kClinicByID(deptID, id), cache.TTLMedium, func() (*model.ClinicDTO, error) {
		return s.repo.GetByID(ctx, deptID, id)
	})
}

func (s *clinicService) List(ctx context.Context, deptID int, q table.TableQuery) (table.TableListResult[model.ClinicDTO], error) {
	type boxed = table.TableListResult[model.ClinicDTO]
	key := kClinicList(deptID, q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.List(ctx, deptID, q)
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

func (s *clinicService) ListByDentistID(ctx context.Context, deptID int, dentistID int, q table.TableQuery) (table.TableListResult[model.ClinicDTO], error) {
	type boxed = table.TableListResult[model.ClinicDTO]
	key := kDentistClinicList(deptID, dentistID, q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.ListByDentistID(ctx, deptID, dentistID, q)
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

func (s *clinicService) ListByPatientID(ctx context.Context, deptID int, dentistID int, q table.TableQuery) (table.TableListResult[model.ClinicDTO], error) {
	type boxed = table.TableListResult[model.ClinicDTO]
	key := kPatientClinicList(deptID, dentistID, q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.ListByPatientID(ctx, deptID, dentistID, q)
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

func (s *clinicService) Delete(ctx context.Context, deptID int, id int) error {
	if err := s.repo.Delete(ctx, deptID, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kClinicAll(deptID)...)
	cache.InvalidateKeys(kClinicByID(deptID, id), kClinicDentistList(id), kClinicPatientList(id))

	s.unlinkSearch(id)
	return nil
}

func (s *clinicService) Search(ctx context.Context, deptID int, q dbutils.SearchQuery) (dbutils.SearchResult[model.ClinicDTO], error) {
	type boxed = dbutils.SearchResult[model.ClinicDTO]
	key := kClinicSearch(deptID, q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.Search(ctx, deptID, q)
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
