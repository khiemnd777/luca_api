
package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/process/repository"
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

type ProcessService interface {
	Create(ctx context.Context, deptID int, input model.ProcessDTO) (*model.ProcessDTO, error)
	Update(ctx context.Context, deptID int, input model.ProcessDTO) (*model.ProcessDTO, error)
	GetByID(ctx context.Context, id int) (*model.ProcessDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.ProcessDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.ProcessDTO], error)
	Delete(ctx context.Context, id int) error
}

type processService struct {
	repo  repository.ProcessRepository
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewProcessService(repo repository.ProcessRepository, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) ProcessService {
	return &processService{repo: repo, deps: deps, cfMgr: cfMgr}
}

// ----------------------------------------------------------------------------
// Cache Keys
// ----------------------------------------------------------------------------

func kProcessByID(id int) string {
	return fmt.Sprintf("process:id:%d", id)
}

func kProcessAll() []string {
	return []string{
		kProcessListAll(),
		kProcessSearchAll(),
	}
}

func kProcessListAll() string {
	return "process:list:*"
}

func kProcessSearchAll() string {
	return "process:search:*"
}

func kProcessList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("process:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func kProcessSearch(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("process:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

// ----------------------------------------------------------------------------
// Create
// ----------------------------------------------------------------------------

func (s *processService) Create(ctx context.Context, deptID int, input model.ProcessDTO) (*model.ProcessDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kProcessByID(dto.ID))
	}
	cache.InvalidateKeys(kProcessAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// Update
// ----------------------------------------------------------------------------

func (s *processService) Update(ctx context.Context, deptID int, input model.ProcessDTO) (*model.ProcessDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kProcessByID(dto.ID))
	}
	cache.InvalidateKeys(kProcessAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// upsertSearch
// ----------------------------------------------------------------------------

func (s *processService) upsertSearch(ctx context.Context, deptID int, dto *model.ProcessDTO) {
	// Bạn có thể chỉnh lại cho phù hợp với module thực tế (Title/Content/Keywords...).
	kwPtr, _ := searchutils.BuildKeywords(ctx, s.cfMgr, "process", []any{dto.Code}, dto.CustomFields)

	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "process",
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

func (s *processService) GetByID(ctx context.Context, id int) (*model.ProcessDTO, error) {
	return cache.Get(kProcessByID(id), cache.TTLMedium, func() (*model.ProcessDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

// ----------------------------------------------------------------------------
// List
// ----------------------------------------------------------------------------

func (s *processService) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.ProcessDTO], error) {
	type boxed = table.TableListResult[model.ProcessDTO]
	key := kProcessList(q)

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

func (s *processService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kProcessAll()...)
	cache.InvalidateKeys(kProcessByID(id))
	return nil
}

// ----------------------------------------------------------------------------
// Search
// ----------------------------------------------------------------------------

func (s *processService) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.ProcessDTO], error) {
	type boxed = dbutils.SearchResult[model.ProcessDTO]
	key := kProcessSearch(q)

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
