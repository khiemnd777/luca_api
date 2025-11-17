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
	GetByID(ctx context.Context, id int) (*model.ClinicDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.ClinicDTO], error)
	ListByDentistID(ctx context.Context, dentistID int, query table.TableQuery) (table.TableListResult[model.ClinicDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.ClinicDTO], error)
	Delete(ctx context.Context, id int) error
}

type clinicService struct {
	repo  repository.ClinicRepository
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewClinicService(repo repository.ClinicRepository, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) ClinicService {
	return &clinicService{repo: repo, deps: deps, cfMgr: cfMgr}
}

func kClinicByID(id int) string {
	return fmt.Sprintf("clinic:id:%d", id)
}

func kClinicAll() []string {
	return []string{
		kClinicListAll(),
		kClinicSearchAll(),
		kClinicDentistAll(),
	}
}

func kClinicListAll() string {
	return "clinic:list:*"
}

func kClinicSearchAll() string {
	return "clinic:search:*"
}

func kClinicDentistAll() string {
	return "clinic:dentist:*"
}

func kClinicDentistList(clinicID int) string {
	return fmt.Sprintf("dentist:clinic:%d:*", clinicID)
}

func kClinicList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("clinic:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func kDentistClinicList(dentistID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("clinic:dentist:%d:list:l%d:p%d:o%s:d%s", dentistID, q.Limit, q.Page, orderBy, q.Direction)
}

func kClinicSearch(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("clinic:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

func (s *clinicService) Create(ctx context.Context, deptID int, input model.ClinicDTO) (*model.ClinicDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kClinicByID(dto.ID), kClinicDentistList(dto.ID))
	}
	cache.InvalidateKeys(kClinicAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

func (s *clinicService) Update(ctx context.Context, deptID int, input model.ClinicDTO) (*model.ClinicDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kClinicByID(dto.ID), kClinicDentistList(dto.ID))
	}
	cache.InvalidateKeys(kClinicAll()...)

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

func (s *clinicService) GetByID(ctx context.Context, id int) (*model.ClinicDTO, error) {
	return cache.Get(kClinicByID(id), cache.TTLMedium, func() (*model.ClinicDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

func (s *clinicService) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.ClinicDTO], error) {
	type boxed = table.TableListResult[model.ClinicDTO]
	key := kClinicList(q)

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

func (s *clinicService) ListByDentistID(ctx context.Context, dentistID int, q table.TableQuery) (table.TableListResult[model.ClinicDTO], error) {
	type boxed = table.TableListResult[model.ClinicDTO]
	key := kDentistClinicList(dentistID, q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.ListByDentistID(ctx, dentistID, q)
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

func (s *clinicService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kClinicAll()...)
	cache.InvalidateKeys(kClinicByID(id), kClinicDentistList(id))
	return nil
}

func (s *clinicService) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.ClinicDTO], error) {
	type boxed = dbutils.SearchResult[model.ClinicDTO]
	key := kClinicSearch(q)

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
