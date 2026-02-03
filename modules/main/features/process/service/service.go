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
	GetByID(ctx context.Context, deptID int, id int) (*model.ProcessDTO, error)
	List(ctx context.Context, deptID int, query table.TableQuery) (table.TableListResult[model.ProcessDTO], error)
	Search(ctx context.Context, deptID int, query dbutils.SearchQuery) (dbutils.SearchResult[model.ProcessDTO], error)
	Delete(ctx context.Context, deptID int, id int) error
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

func kProcessByID(deptID int, id int) string {
	return fmt.Sprintf("process:dpt%d:id:%d", deptID, id)
}

func kProcessAll(deptID int) []string {
	return []string{
		kProcessListAll(deptID),
		kProcessSearchAll(deptID),
		kProcessSectionAll(deptID),
		"product_process:list:*",
		"category_process:list:*",
	}
}

func kProcessListAll(deptID int) string {
	return fmt.Sprintf("process:list:dpt%d:*", deptID)
}

func kProcessSearchAll(deptID int) string {
	return fmt.Sprintf("process:search:dpt%d:*", deptID)
}

func kProcessSectionAll(deptID int) string {
	return fmt.Sprintf("process:section:dpt%d:*", deptID)
}

func kProcessList(deptID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("process:list:dpt%d:l%d:p%d:o%s:d%s", deptID, q.Limit, q.Page, orderBy, q.Direction)
}

func kSectionProcessesList(deptID int, sectionID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("section:dpt%d:id:%d:processes:list:l%d:p%d:o%s:d%s", deptID, sectionID, q.Limit, q.Page, orderBy, q.Direction)
}

func kProcessSearch(deptID int, q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("process:search:dpt%d:k%s:l%d:p%d:o%s:d%s", deptID, q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

// ----------------------------------------------------------------------------
// Create
// ----------------------------------------------------------------------------

func (s *processService) Create(ctx context.Context, deptID int, input model.ProcessDTO) (*model.ProcessDTO, error) {
	dto, err := s.repo.Create(ctx, deptID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kProcessByID(deptID, dto.ID))
	}
	cache.InvalidateKeys(kProcessAll(deptID)...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// Update
// ----------------------------------------------------------------------------

func (s *processService) Update(ctx context.Context, deptID int, input model.ProcessDTO) (*model.ProcessDTO, error) {
	dto, err := s.repo.Update(ctx, deptID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kProcessByID(deptID, dto.ID))
	}
	cache.InvalidateKeys(kProcessAll(deptID)...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// upsertSearch
// ----------------------------------------------------------------------------

func (s *processService) upsertSearch(ctx context.Context, deptID int, dto *model.ProcessDTO) {
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

func (s *processService) unlinkSearch(id int) {
	pubsub.PublishAsync("search:unlink", &searchmodel.UnlinkDoc{
		EntityType: "process",
		EntityID:   int64(id),
	})
}

// ----------------------------------------------------------------------------
// GetByID
// ----------------------------------------------------------------------------

func (s *processService) GetByID(ctx context.Context, deptID int, id int) (*model.ProcessDTO, error) {
	return cache.Get(kProcessByID(deptID, id), cache.TTLMedium, func() (*model.ProcessDTO, error) {
		return s.repo.GetByID(ctx, deptID, id)
	})
}

// ----------------------------------------------------------------------------
// List
// ----------------------------------------------------------------------------

func (s *processService) List(ctx context.Context, deptID int, q table.TableQuery) (table.TableListResult[model.ProcessDTO], error) {
	type boxed = table.TableListResult[model.ProcessDTO]
	key := kProcessList(deptID, q)

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

func (s *processService) Delete(ctx context.Context, deptID int, id int) error {
	_, err := s.repo.GetByID(ctx, deptID, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, deptID, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kProcessAll(deptID)...)
	cache.InvalidateKeys(kProcessByID(deptID, id))

	s.unlinkSearch(id)
	return nil
}

// ----------------------------------------------------------------------------
// Search
// ----------------------------------------------------------------------------

func (s *processService) Search(ctx context.Context, deptID int, q dbutils.SearchQuery) (dbutils.SearchResult[model.ProcessDTO], error) {
	type boxed = dbutils.SearchResult[model.ProcessDTO]
	key := kProcessSearch(deptID, q)

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
