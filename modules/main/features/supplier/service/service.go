package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/supplier/repository"
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

type SupplierService interface {
	Create(ctx context.Context, deptID int, input model.SupplierDTO) (*model.SupplierDTO, error)
	Update(ctx context.Context, deptID int, input model.SupplierDTO) (*model.SupplierDTO, error)
	GetByID(ctx context.Context, id int) (*model.SupplierDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.SupplierDTO], error)
	ListByMaterialID(ctx context.Context, materialID int, query table.TableQuery) (table.TableListResult[model.SupplierDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.SupplierDTO], error)
	Delete(ctx context.Context, id int) error
}

type supplierService struct {
	repo  repository.SupplierRepository
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewSupplierService(repo repository.SupplierRepository, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) SupplierService {
	return &supplierService{repo: repo, deps: deps, cfMgr: cfMgr}
}

// ----------------------------------------------------------------------------
// Cache Keys
// ----------------------------------------------------------------------------

func kSupplierByID(id int) string {
	return fmt.Sprintf("supplier:id:%d", id)
}

func kSupplierAll() []string {
	return []string{
		kSupplierListAll(),
		kSupplierSearchAll(),
	}
}

func kSupplierListAll() string {
	return "supplier:list:*"
}

func kSupplierSearchAll() string {
	return "supplier:search:*"
}

func kSupplierMaterialList(materialID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("supplier:material:%d:list:l%d:p%d:o%s:d%s", materialID, q.Limit, q.Page, orderBy, q.Direction)
}

func kSupplierList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("supplier:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func kSupplierSearch(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("supplier:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

// ----------------------------------------------------------------------------
// Create
// ----------------------------------------------------------------------------

func (s *supplierService) Create(ctx context.Context, deptID int, input model.SupplierDTO) (*model.SupplierDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kSupplierByID(dto.ID))
	}
	cache.InvalidateKeys(kSupplierAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// Update
// ----------------------------------------------------------------------------

func (s *supplierService) Update(ctx context.Context, deptID int, input model.SupplierDTO) (*model.SupplierDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kSupplierByID(dto.ID))
	}
	cache.InvalidateKeys(kSupplierAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// upsertSearch
// ----------------------------------------------------------------------------

func (s *supplierService) upsertSearch(ctx context.Context, deptID int, dto *model.SupplierDTO) {
	// Bạn có thể chỉnh lại cho phù hợp với module thực tế (Title/Content/Keywords...).
	kwPtr, _ := searchutils.BuildKeywords(ctx, s.cfMgr, "supplier", []any{dto.Code}, dto.CustomFields)

	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "supplier",
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

func (s *supplierService) GetByID(ctx context.Context, id int) (*model.SupplierDTO, error) {
	return cache.Get(kSupplierByID(id), cache.TTLMedium, func() (*model.SupplierDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

// ----------------------------------------------------------------------------
// List
// ----------------------------------------------------------------------------

func (s *supplierService) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.SupplierDTO], error) {
	type boxed = table.TableListResult[model.SupplierDTO]
	key := kSupplierList(q)

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

func (s *supplierService) ListByMaterialID(ctx context.Context, materialID int, query table.TableQuery) (table.TableListResult[model.SupplierDTO], error) {
	type boxed = table.TableListResult[model.SupplierDTO]
	key := kSupplierMaterialList(materialID, query)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.ListByMaterialID(ctx, materialID, query)
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

func (s *supplierService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kSupplierAll()...)
	cache.InvalidateKeys(kSupplierByID(id))
	return nil
}

// ----------------------------------------------------------------------------
// Search
// ----------------------------------------------------------------------------

func (s *supplierService) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.SupplierDTO], error) {
	type boxed = dbutils.SearchResult[model.SupplierDTO]
	key := kSupplierSearch(q)

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
