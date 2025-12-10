package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/category/repository"
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

type CategoryService interface {
	Create(ctx context.Context, deptID int, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error)
	Update(ctx context.Context, deptID int, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error)
	GetByID(ctx context.Context, id int) (*model.CategoryDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.CategoryDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.CategoryDTO], error)
	Delete(ctx context.Context, id int) error
}

type categoryService struct {
	repo  repository.CategoryRepository
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewCategoryService(repo repository.CategoryRepository, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) CategoryService {
	return &categoryService{repo: repo, deps: deps, cfMgr: cfMgr}
}

// ----------------------------------------------------------------------------
// Cache Keys
// ----------------------------------------------------------------------------

func kCategoryByID(id int) string {
	return fmt.Sprintf("category:id:%d", id)
}

func kCategoryAll() []string {
	return []string{
		kCategoryListAll(),
		kCategorySearchAll(),
	}
}

func kCategoryListAll() string {
	return "category:list:*"
}

func kCategorySearchAll() string {
	return "category:search:*"
}

func kCategoryList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("category:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func kCategorySearch(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("category:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

// ----------------------------------------------------------------------------
// Create
// ----------------------------------------------------------------------------

func (s *categoryService) Create(ctx context.Context, deptID int, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kCategoryByID(dto.ID))
	}
	cache.InvalidateKeys(kCategoryAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// Update
// ----------------------------------------------------------------------------

func (s *categoryService) Update(ctx context.Context, deptID int, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kCategoryByID(dto.ID))
	}
	cache.InvalidateKeys(kCategoryAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// upsertSearch
// ----------------------------------------------------------------------------

func (s *categoryService) upsertSearch(ctx context.Context, deptID int, dto *model.CategoryDTO) {
	kwPtr, _ := searchutils.BuildKeywords(ctx, s.cfMgr, "category", []any{dto.Name}, dto.CustomFields)

	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "category",
		EntityID:   int64(dto.ID),
		Title:      *dto.Name,
		Subtitle:   nil,
		Keywords:   &kwPtr,
		Content:    nil,
		Attributes: map[string]any{},
		OrgID:      utils.Ptr(int64(deptID)),
		OwnerID:    nil,
	})
}

// ----------------------------------------------------------------------------
// GetByID
// ----------------------------------------------------------------------------

func (s *categoryService) GetByID(ctx context.Context, id int) (*model.CategoryDTO, error) {
	return cache.Get(kCategoryByID(id), cache.TTLMedium, func() (*model.CategoryDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

// ----------------------------------------------------------------------------
// List
// ----------------------------------------------------------------------------

func (s *categoryService) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.CategoryDTO], error) {
	type boxed = table.TableListResult[model.CategoryDTO]
	key := kCategoryList(q)

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

// ----------------------------------------------------------------------------
// Delete
// ----------------------------------------------------------------------------

func (s *categoryService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kCategoryAll()...)
	cache.InvalidateKeys(kCategoryByID(id))
	return nil
}

// ----------------------------------------------------------------------------
// Search
// ----------------------------------------------------------------------------

func (s *categoryService) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.CategoryDTO], error) {
	type boxed = dbutils.SearchResult[model.CategoryDTO]
	key := kCategorySearch(q)

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
