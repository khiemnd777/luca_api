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
	GetByID(ctx context.Context, deptID int, id int) (*model.CategoryDTO, error)
	List(ctx context.Context, deptID int, query table.TableQuery) (table.TableListResult[model.CategoryDTO], error)
	Search(ctx context.Context, deptID int, query dbutils.SearchQuery) (dbutils.SearchResult[model.CategoryDTO], error)
	Delete(ctx context.Context, deptID int, id int) error
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

func kCategoryByID(deptID int, id int) string {
	return fmt.Sprintf("category:dpt%d:id:%d", deptID, id)
}

func kCategoryAll(deptID int) []string {
	return []string{
		kCategoryListAll(deptID),
		kCategorySearchAll(deptID),
		"collections:list:g=category:*",
	}
}

func kCategoryListAll(deptID int) string {
	return fmt.Sprintf("category:list:dpt%d:*", deptID)
}

func kCategorySearchAll(deptID int) string {
	return fmt.Sprintf("category:search:dpt%d:*", deptID)
}

func kCategoryList(deptID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("category:list:dpt%d:l%d:p%d:o%s:d%s", deptID, q.Limit, q.Page, orderBy, q.Direction)
}

func kCategorySearch(deptID int, q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("category:search:dpt%d:k%s:l%d:p%d:o%s:d%s", deptID, q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

// ----------------------------------------------------------------------------
// Create
// ----------------------------------------------------------------------------

func (s *categoryService) Create(ctx context.Context, deptID int, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error) {
	dto, err := s.repo.Create(ctx, deptID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kCategoryByID(deptID, dto.ID))
	}
	cache.InvalidateKeys(kCategoryAll(deptID)...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// Update
// ----------------------------------------------------------------------------

func (s *categoryService) Update(ctx context.Context, deptID int, input *model.CategoryUpsertDTO) (*model.CategoryDTO, error) {
	dto, err := s.repo.Update(ctx, deptID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(
			kCategoryByID(deptID, dto.ID),
		)
	}
	cache.InvalidateKeys(kCategoryAll(deptID)...)

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

func (s *categoryService) unlinkSearch(id int) {
	pubsub.PublishAsync("search:unlink", &searchmodel.UnlinkDoc{
		EntityType: "category",
		EntityID:   int64(id),
	})
}

// ----------------------------------------------------------------------------
// GetByID
// ----------------------------------------------------------------------------

func (s *categoryService) GetByID(ctx context.Context, deptID int, id int) (*model.CategoryDTO, error) {
	return cache.Get(kCategoryByID(deptID, id), cache.TTLMedium, func() (*model.CategoryDTO, error) {
		return s.repo.GetByID(ctx, deptID, id)
	})
}

// ----------------------------------------------------------------------------
// List
// ----------------------------------------------------------------------------

func (s *categoryService) List(ctx context.Context, deptID int, q table.TableQuery) (table.TableListResult[model.CategoryDTO], error) {
	type boxed = table.TableListResult[model.CategoryDTO]
	key := kCategoryList(deptID, q)

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

// ----------------------------------------------------------------------------
// Delete
// ----------------------------------------------------------------------------

func (s *categoryService) Delete(ctx context.Context, deptID int, id int) error {
	if err := s.repo.Delete(ctx, deptID, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kCategoryAll(deptID)...)
	cache.InvalidateKeys(kCategoryByID(deptID, id))

	s.unlinkSearch(id)
	return nil
}

// ----------------------------------------------------------------------------
// Search
// ----------------------------------------------------------------------------

func (s *categoryService) Search(ctx context.Context, deptID int, q dbutils.SearchQuery) (dbutils.SearchResult[model.CategoryDTO], error) {
	type boxed = dbutils.SearchResult[model.CategoryDTO]
	key := kCategorySearch(deptID, q)

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
