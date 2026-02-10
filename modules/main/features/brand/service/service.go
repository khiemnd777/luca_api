package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/brand/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/module"
	searchmodel "github.com/khiemnd777/andy_api/shared/modules/search/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type BrandNameService interface {
	Create(ctx context.Context, deptID int, input model.BrandNameDTO) (*model.BrandNameDTO, error)
	Update(ctx context.Context, deptID int, input model.BrandNameDTO) (*model.BrandNameDTO, error)
	GetByID(ctx context.Context, deptID int, id int) (*model.BrandNameDTO, error)
	List(ctx context.Context, deptID int, categoryID *int, query table.TableQuery) (table.TableListResult[model.BrandNameDTO], error)
	Search(ctx context.Context, deptID int, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.BrandNameDTO], error)
	Delete(ctx context.Context, deptID int, id int) error
}

type brandNameService struct {
	repo repository.BrandNameRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewBrandNameService(repo repository.BrandNameRepository, deps *module.ModuleDeps[config.ModuleConfig]) BrandNameService {
	return &brandNameService{repo: repo, deps: deps}
}

func kBrandNameByID(deptID int, id int) string {
	return fmt.Sprintf("brand:name:dpt%d:id:%d", deptID, id)
}

func kBrandNameAll(deptID int) []string {
	return []string{
		kBrandNameListAll(deptID),
		kBrandNameSearchAll(deptID),
	}
}

func kBrandNameListAll(deptID int) string {
	return fmt.Sprintf("brand:name:list:dpt%d:*", deptID)
}

func kBrandNameSearchAll(deptID int) string {
	return fmt.Sprintf("brand:name:search:dpt%d:*", deptID)
}

func kBrandNameList(deptID int, categoryID *int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	cid := 0
	if categoryID != nil {
		cid = *categoryID
	}
	return fmt.Sprintf("brand:name:list:dpt%d:c%d:l%d:p%d:o%s:d%s", deptID, cid, q.Limit, q.Page, orderBy, q.Direction)
}

func kBrandNameSearch(deptID int, categoryID *int, q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	cid := 0
	if categoryID != nil {
		cid = *categoryID
	}
	return fmt.Sprintf("brand:name:search:dpt%d:c%d:k%s:l%d:p%d:o%s:d%s", deptID, cid, q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

func (s *brandNameService) Create(ctx context.Context, deptID int, input model.BrandNameDTO) (*model.BrandNameDTO, error) {
	dto, err := s.repo.Create(ctx, deptID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kBrandNameByID(deptID, dto.ID))
	}
	cache.InvalidateKeys(kBrandNameAll(deptID)...)

	s.upsertSearch(deptID, dto)

	return dto, nil
}

func (s *brandNameService) Update(ctx context.Context, deptID int, input model.BrandNameDTO) (*model.BrandNameDTO, error) {
	dto, err := s.repo.Update(ctx, deptID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kBrandNameByID(deptID, dto.ID))
	}
	cache.InvalidateKeys(kBrandNameAll(deptID)...)

	s.upsertSearch(deptID, dto)

	return dto, nil
}

func (s *brandNameService) upsertSearch(deptID int, dto *model.BrandNameDTO) {
	if dto == nil || dto.Name == nil {
		return
	}
	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "brand",
		EntityID:   int64(dto.ID),
		Title:      *dto.Name,
		Subtitle:   nil,
		Keywords:   dto.Name,
		Content:    nil,
		Attributes: nil,
		OrgID:      utils.Ptr(int64(deptID)),
		OwnerID:    nil,
	})
}

func (s *brandNameService) unlinkSearch(id int) {
	pubsub.PublishAsync("search:unlink", &searchmodel.UnlinkDoc{
		EntityType: "brand",
		EntityID:   int64(id),
	})
}

func (s *brandNameService) GetByID(ctx context.Context, deptID int, id int) (*model.BrandNameDTO, error) {
	return cache.Get(kBrandNameByID(deptID, id), cache.TTLMedium, func() (*model.BrandNameDTO, error) {
		return s.repo.GetByID(ctx, deptID, id)
	})
}

func (s *brandNameService) List(ctx context.Context, deptID int, categoryID *int, q table.TableQuery) (table.TableListResult[model.BrandNameDTO], error) {
	type boxed = table.TableListResult[model.BrandNameDTO]
	key := kBrandNameList(deptID, categoryID, q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.List(ctx, deptID, categoryID, q)
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

func (s *brandNameService) Delete(ctx context.Context, deptID int, id int) error {
	if err := s.repo.Delete(ctx, deptID, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kBrandNameAll(deptID)...)
	cache.InvalidateKeys(kBrandNameByID(deptID, id))

	s.unlinkSearch(id)
	return nil
}

func (s *brandNameService) Search(ctx context.Context, deptID int, categoryID *int, q dbutils.SearchQuery) (dbutils.SearchResult[model.BrandNameDTO], error) {
	type boxed = dbutils.SearchResult[model.BrandNameDTO]
	key := kBrandNameSearch(deptID, categoryID, q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.Search(ctx, deptID, categoryID, q)
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
