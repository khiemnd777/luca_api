package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/product/repository"
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

type ProductService interface {
	Create(ctx context.Context, deptID int, input *model.ProductUpsertDTO) (*model.ProductDTO, error)
	Update(ctx context.Context, deptID int, input *model.ProductUpsertDTO) (*model.ProductDTO, error)
	GetByID(ctx context.Context, id int) (*model.ProductDTO, error)
	List(ctx context.Context, query table.TableQuery) (table.TableListResult[model.ProductDTO], error)
	VariantList(ctx context.Context, templateID int, query table.TableQuery) (table.TableListResult[model.ProductDTO], error)
	Search(ctx context.Context, query dbutils.SearchQuery) (dbutils.SearchResult[model.ProductDTO], error)
	Delete(ctx context.Context, id int) error
}

type productService struct {
	repo  repository.ProductRepository
	deps  *module.ModuleDeps[config.ModuleConfig]
	cfMgr *customfields.Manager
}

func NewProductService(repo repository.ProductRepository, deps *module.ModuleDeps[config.ModuleConfig], cfMgr *customfields.Manager) ProductService {
	return &productService{repo: repo, deps: deps, cfMgr: cfMgr}
}

// ----------------------------------------------------------------------------
// Cache Keys
// ----------------------------------------------------------------------------

func kProductByID(id int) string {
	return fmt.Sprintf("product:id:%d", id)
}

func kProductAll() []string {
	return []string{
		kProductListAll(),
		kProductSearchAll(),
	}
}

func kProductListAll() string {
	return "product:list:*"
}

func kProductSearchAll() string {
	return "product:search:*"
}

func kProductList(q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("product:list:l%d:p%d:o%s:d%s", q.Limit, q.Page, orderBy, q.Direction)
}

func kVariantProductList(templateID int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("product:list:vid:%d:l%d:p%d:o%s:d%s", templateID, q.Limit, q.Page, orderBy, q.Direction)
}

func kProductSearch(q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	return fmt.Sprintf("product:search:k%s:l%d:p%d:o%s:d%s", q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

// ----------------------------------------------------------------------------
// Create
// ----------------------------------------------------------------------------

func (s *productService) Create(ctx context.Context, deptID int, input *model.ProductUpsertDTO) (*model.ProductDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kProductByID(dto.ID))
	}
	cache.InvalidateKeys(kProductAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// Update
// ----------------------------------------------------------------------------

func (s *productService) Update(ctx context.Context, deptID int, input *model.ProductUpsertDTO) (*model.ProductDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kProductByID(dto.ID))
	}
	cache.InvalidateKeys(kProductAll()...)

	s.upsertSearch(ctx, deptID, dto)

	return dto, nil
}

// ----------------------------------------------------------------------------
// upsertSearch
// ----------------------------------------------------------------------------

func (s *productService) upsertSearch(ctx context.Context, deptID int, dto *model.ProductDTO) {
	kwPtr, _ := searchutils.BuildKeywords(ctx, s.cfMgr, "product", []any{dto.Code}, dto.CustomFields)

	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "product",
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

func (s *productService) unlinkSearch(id int) {
	pubsub.PublishAsync("search:unlink", &searchmodel.UnlinkDoc{
		EntityType: "product",
		EntityID:   int64(id),
	})
}

// ----------------------------------------------------------------------------
// GetByID
// ----------------------------------------------------------------------------

func (s *productService) GetByID(ctx context.Context, id int) (*model.ProductDTO, error) {
	return cache.Get(kProductByID(id), cache.TTLMedium, func() (*model.ProductDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

// ----------------------------------------------------------------------------
// List
// ----------------------------------------------------------------------------

func (s *productService) List(ctx context.Context, q table.TableQuery) (table.TableListResult[model.ProductDTO], error) {
	type boxed = table.TableListResult[model.ProductDTO]
	key := kProductList(q)

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
// Variant List
// ----------------------------------------------------------------------------

func (s *productService) VariantList(ctx context.Context, templateID int, query table.TableQuery) (table.TableListResult[model.ProductDTO], error) {
	type boxed = table.TableListResult[model.ProductDTO]
	key := kVariantProductList(templateID, query)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.VariantList(ctx, templateID, query)
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

func (s *productService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kProductAll()...)
	cache.InvalidateKeys(kProductByID(id))

	s.unlinkSearch(id)
	return nil
}

// ----------------------------------------------------------------------------
// Search
// ----------------------------------------------------------------------------

func (s *productService) Search(ctx context.Context, q dbutils.SearchQuery) (dbutils.SearchResult[model.ProductDTO], error) {
	type boxed = dbutils.SearchResult[model.ProductDTO]
	key := kProductSearch(q)

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
