package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/section/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/module"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type SectionService interface {
	Create(ctx context.Context, input model.SectionDTO) (*model.SectionDTO, error)
	Update(ctx context.Context, input model.SectionDTO) (*model.SectionDTO, error)
	GetByID(ctx context.Context, id int) (*model.SectionDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.SectionDTO], error)
	ListByStaffID(ctx context.Context, staffID int, query table.TableQuery) (table.TableListResult[model.SectionDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.SectionDTO], error)
	Delete(ctx context.Context, id int) error
}

type sectionService struct {
	repo repository.SectionRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewSectionService(repo repository.SectionRepository, deps *module.ModuleDeps[config.ModuleConfig]) SectionService {
	return &sectionService{repo: repo, deps: deps}
}

func kSectionByID(id int) string {
	return fmt.Sprintf("section:id:%d", id)
}

func kSectionAll() []string {
	return []string{kSectionListAll(), kSectionSearchAll()}
}

func kSectionListAll() string {
	return "section:list:*"
}

func kSectionSearchAll() string {
	return "section:search:*"
}

func kSectionList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("section:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func kSectionStaffList(staffID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("section:staff:%d:list:l%d:p%d:o%s:d%s", staffID, q.Limit, q.Page, orderBy, q.Direction)
}

func kSectionSearch(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("section:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

func (s *sectionService) Create(ctx context.Context, input model.SectionDTO) (*model.SectionDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	cache.InvalidateKeys(kSectionAll()...)
	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kSectionByID(dto.ID))
	}
	return dto, nil
}

func (s *sectionService) Update(ctx context.Context, input model.SectionDTO) (*model.SectionDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kSectionByID(dto.ID))
	}
	cache.InvalidateKeys(kSectionAll()...)
	return dto, nil
}

func (s *sectionService) GetByID(ctx context.Context, id int) (*model.SectionDTO, error) {
	return cache.Get(kSectionByID(id), cache.TTLMedium, func() (*model.SectionDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

func (s *sectionService) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.SectionDTO], error) {
	type boxed = table.TableListResult[model.SectionDTO]
	key := kSectionList(q)

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

func (s *sectionService) ListByStaffID(ctx context.Context, staffID int, q table.TableQuery) (table.TableListResult[model.SectionDTO], error) {
	type boxed = table.TableListResult[model.SectionDTO]
	key := kSectionStaffList(staffID, q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.ListByStaffID(ctx, staffID, q)
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

func (s *sectionService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kSectionAll()...)
	cache.InvalidateKeys(kSectionByID(id))
	return nil
}

func (s *sectionService) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.SectionDTO], error) {
	type boxed = dbutils.SearchResult[model.SectionDTO]
	key := kSectionSearch(q)

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
