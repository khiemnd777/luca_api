
package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/customer/repository"
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

type CustomerService interface {
	Create(ctx context.Context, deptID int, input model.CustomerDTO) (*model.CustomerDTO, error)
	Update(ctx context.Context, deptID int, input model.CustomerDTO) (*model.CustomerDTO, error)
	GetByID(ctx context.Context, id int) (*model.CustomerDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.CustomerDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.CustomerDTO], error)
	Delete(ctx context.Context, id int) error
}

type customerService struct {
	repo  repository.CustomerRepository
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewCustomerService(repo repository.CustomerRepository, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) CustomerService {
	return &customerService{repo: repo, deps: deps, cfMgr: cfMgr}
}

// ----------------------------------------------------------------------------
// Cache Keys
// ----------------------------------------------------------------------------

func kCustomerByID(id int) string {
	return fmt.Sprintf("customer:id:%d", id)
}

func kCustomerAll() []string {
	return []string{
		kCustomerListAll(),
		kCustomerSearchAll(),
	}
}

func kCustomerListAll() string {
	return "customer:list:*"
}

func kCustomerSearchAll() string {
	return "customer:search:*"
}

func kCustomerList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("customer:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func kCustomerSearch(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("customer:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

// ----------------------------------------------------------------------------
// Create
// ----------------------------------------------------------------------------

func (s *customerService) Create(ctx context.Context, deptID int, input model.CustomerDTO) (*model.CustomerDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kCustomerByID(dto.ID))
	}
	cache.InvalidateKeys(kCustomerAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// Update
// ----------------------------------------------------------------------------

func (s *customerService) Update(ctx context.Context, deptID int, input model.CustomerDTO) (*model.CustomerDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kCustomerByID(dto.ID))
	}
	cache.InvalidateKeys(kCustomerAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// upsertSearch
// ----------------------------------------------------------------------------

func (s *customerService) upsertSearch(ctx context.Context, deptID int, dto *model.CustomerDTO) {
	// Bạn có thể chỉnh lại cho phù hợp với module thực tế (Title/Content/Keywords...).
	kwPtr, _ := searchutils.BuildKeywords(ctx, s.cfMgr, "customer", []any{dto.Code}, dto.CustomFields)

	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "customer",
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

func (s *customerService) GetByID(ctx context.Context, id int) (*model.CustomerDTO, error) {
	return cache.Get(kCustomerByID(id), cache.TTLMedium, func() (*model.CustomerDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

// ----------------------------------------------------------------------------
// List
// ----------------------------------------------------------------------------

func (s *customerService) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.CustomerDTO], error) {
	type boxed = table.TableListResult[model.CustomerDTO]
	key := kCustomerList(q)

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

func (s *customerService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kCustomerAll()...)
	cache.InvalidateKeys(kCustomerByID(id))
	return nil
}

// ----------------------------------------------------------------------------
// Search
// ----------------------------------------------------------------------------

func (s *customerService) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.CustomerDTO], error) {
	type boxed = dbutils.SearchResult[model.CustomerDTO]
	key := kCustomerSearch(q)

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
