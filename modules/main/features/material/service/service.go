package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/material/repository"
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

type MaterialService interface {
	Create(ctx context.Context, deptID int, input model.MaterialDTO) (*model.MaterialDTO, error)
	Update(ctx context.Context, deptID int, input model.MaterialDTO) (*model.MaterialDTO, error)
	GetByID(ctx context.Context, id int) (*model.MaterialDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.MaterialDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.MaterialDTO], error)
	Delete(ctx context.Context, id int) error
}

type materialService struct {
	repo  repository.MaterialRepository
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewMaterialService(repo repository.MaterialRepository, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) MaterialService {
	return &materialService{repo: repo, deps: deps, cfMgr: cfMgr}
}

// ----------------------------------------------------------------------------
// Cache Keys
// ----------------------------------------------------------------------------

func kMaterialByID(id int) string {
	return fmt.Sprintf("material:id:%d", id)
}

func kMaterialAll() []string {
	return []string{
		kMaterialListAll(),
		kMaterialSearchAll(),
	}
}

func kMaterialListAll() string {
	return "material:list:*"
}

func kMaterialSearchAll() string {
	return "material:search:*"
}

func kMaterialList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("material:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func kMaterialSearch(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("material:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

// ----------------------------------------------------------------------------
// Create
// ----------------------------------------------------------------------------

func (s *materialService) Create(ctx context.Context, deptID int, input model.MaterialDTO) (*model.MaterialDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kMaterialByID(dto.ID))
	}
	cache.InvalidateKeys(kMaterialAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// Update
// ----------------------------------------------------------------------------

func (s *materialService) Update(ctx context.Context, deptID int, input model.MaterialDTO) (*model.MaterialDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kMaterialByID(dto.ID))
	}
	cache.InvalidateKeys(kMaterialAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// upsertSearch
// ----------------------------------------------------------------------------

func (s *materialService) upsertSearch(ctx context.Context, deptID int, dto *model.MaterialDTO) {
	// Bạn có thể chỉnh lại cho phù hợp với module thực tế (Title/Content/Keywords...).
	kwPtr, _ := searchutils.BuildKeywords(ctx, s.cfMgr, "material", []any{dto.Code}, dto.CustomFields)

	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "material",
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

func (s *materialService) GetByID(ctx context.Context, id int) (*model.MaterialDTO, error) {
	return cache.Get(kMaterialByID(id), cache.TTLMedium, func() (*model.MaterialDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

// ----------------------------------------------------------------------------
// List
// ----------------------------------------------------------------------------

func (s *materialService) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.MaterialDTO], error) {
	type boxed = table.TableListResult[model.MaterialDTO]
	key := kMaterialList(q)

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

func (s *materialService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kMaterialAll()...)
	cache.InvalidateKeys(kMaterialByID(id))
	return nil
}

// ----------------------------------------------------------------------------
// Search
// ----------------------------------------------------------------------------

func (s *materialService) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.MaterialDTO], error) {
	type boxed = dbutils.SearchResult[model.MaterialDTO]
	key := kMaterialSearch(q)

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
