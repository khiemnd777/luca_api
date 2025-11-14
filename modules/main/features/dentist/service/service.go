package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/dentist/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/module"
	searchmodel "github.com/khiemnd777/andy_api/shared/modules/search/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type DentistService interface {
	Create(ctx context.Context, deptID int, input model.DentistDTO) (*model.DentistDTO, error)
	Update(ctx context.Context, deptID int, input model.DentistDTO) (*model.DentistDTO, error)
	GetByID(ctx context.Context, id int) (*model.DentistDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.DentistDTO], error)
	ListByClinicID(ctx context.Context, clinicID int, query table.TableQuery) (table.TableListResult[model.DentistDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.DentistDTO], error)
	Delete(ctx context.Context, id int) error
}

type dentistService struct {
	repo repository.DentistRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewDentistService(repo repository.DentistRepository, deps *module.ModuleDeps[config.ModuleConfig]) DentistService {
	return &dentistService{repo: repo, deps: deps}
}

func kDentistByID(id int) string {
	return fmt.Sprintf("dentist:id:%d", id)
}

func kDentistAll() []string {
	return []string{
		kDentistListAll(),
		kDentistSearchAll(),
		kDentistClinicAll(),
	}
}

func kDentistListAll() string {
	return "dentist:list:*"
}

func kDentistSearchAll() string {
	return "dentist:search:*"
}

func kDentistClinicAll() string {
	return "dentist:clinic:*"
}

func kDentistClinicList(dentistID int) string {
	return fmt.Sprintf("clinic:dentist:%d:*", dentistID)
}

func kDentistList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("dentist:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func kClinicDentistList(clinicID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("dentist:clinic:%d:list:l%d:p%d:o%s:d%s", clinicID, q.Limit, q.Page, orderBy, q.Direction)
}

func kDentistSearch(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("dentist:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

func (s *dentistService) Create(ctx context.Context, deptID int, input model.DentistDTO) (*model.DentistDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kDentistByID(dto.ID), kDentistClinicList(dto.ID))
	}

	cache.InvalidateKeys(kDentistAll()...)

	upsertSearch(deptID, dto)

	return dto, nil
}

func (s *dentistService) Update(ctx context.Context, deptID int, input model.DentistDTO) (*model.DentistDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kDentistByID(dto.ID), kDentistClinicList(dto.ID))
	}
	cache.InvalidateKeys(kDentistAll()...)

	upsertSearch(deptID, dto)

	return dto, nil
}

func upsertSearch(deptID int, dto *model.DentistDTO) {
	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "dentist",
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

func (s *dentistService) GetByID(ctx context.Context, id int) (*model.DentistDTO, error) {
	return cache.Get(kDentistByID(id), cache.TTLMedium, func() (*model.DentistDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

func (s *dentistService) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.DentistDTO], error) {
	type boxed = table.TableListResult[model.DentistDTO]
	key := kDentistList(q)

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

func (s *dentistService) ListByClinicID(ctx context.Context, clinicID int, query table.TableQuery) (table.TableListResult[model.DentistDTO], error) {
	type boxed = table.TableListResult[model.DentistDTO]
	key := kClinicDentistList(clinicID, query)

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

func (s *dentistService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kDentistAll()...)
	cache.InvalidateKeys(kDentistByID(id), kDentistClinicList(id))
	return nil
}

func (s *dentistService) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.DentistDTO], error) {
	type boxed = dbutils.SearchResult[model.DentistDTO]
	key := kDentistSearch(q)

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
