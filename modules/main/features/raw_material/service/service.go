package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/raw_material/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/module"
	searchmodel "github.com/khiemnd777/andy_api/shared/modules/search/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type RawMaterialService interface {
	Create(ctx context.Context, deptID int, input model.RawMaterialDTO) (*model.RawMaterialDTO, error)
	Update(ctx context.Context, deptID int, input model.RawMaterialDTO) (*model.RawMaterialDTO, error)
	GetByID(ctx context.Context, deptID int, id int) (*model.RawMaterialDTO, error)
	List(ctx context.Context, deptID int, categoryID *int, query table.TableQuery) (table.TableListResult[model.RawMaterialDTO], error)
	Search(ctx context.Context, deptID int, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.RawMaterialDTO], error)
	Delete(ctx context.Context, deptID int, id int) error
}

type rawMaterialService struct {
	repo repository.RawMaterialRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewRawMaterialService(repo repository.RawMaterialRepository, deps *module.ModuleDeps[config.ModuleConfig]) RawMaterialService {
	return &rawMaterialService{repo: repo, deps: deps}
}

func kRawMaterialByID(deptID int, id int) string {
	return fmt.Sprintf("raw_material:dpt%d:id:%d", deptID, id)
}

func kRawMaterialAll(deptID int) []string {
	return []string{
		kRawMaterialListAll(deptID),
		kRawMaterialSearchAll(deptID),
	}
}

func kRawMaterialListAll(deptID int) string {
	return fmt.Sprintf("raw_material:list:dpt%d:*", deptID)
}

func kRawMaterialSearchAll(deptID int) string {
	return fmt.Sprintf("raw_material:search:dpt%d:*", deptID)
}

func kRawMaterialList(deptID int, categoryID *int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	cid := 0
	if categoryID != nil {
		cid = *categoryID
	}
	return fmt.Sprintf("raw_material:list:dpt%d:c%d:l%d:p%d:o%s:d%s", deptID, cid, q.Limit, q.Page, orderBy, q.Direction)
}

func kRawMaterialSearch(deptID int, categoryID *int, q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	cid := 0
	if categoryID != nil {
		cid = *categoryID
	}
	return fmt.Sprintf("raw_material:search:dpt%d:c%d:k%s:l%d:p%d:o%s:d%s", deptID, cid, q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

func (s *rawMaterialService) Create(ctx context.Context, deptID int, input model.RawMaterialDTO) (*model.RawMaterialDTO, error) {
	dto, err := s.repo.Create(ctx, deptID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kRawMaterialByID(deptID, dto.ID))
	}
	cache.InvalidateKeys(kRawMaterialAll(deptID)...)

	s.upsertSearch(deptID, dto)

	return dto, nil
}

func (s *rawMaterialService) Update(ctx context.Context, deptID int, input model.RawMaterialDTO) (*model.RawMaterialDTO, error) {
	dto, err := s.repo.Update(ctx, deptID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kRawMaterialByID(deptID, dto.ID))
	}
	cache.InvalidateKeys(kRawMaterialAll(deptID)...)

	s.upsertSearch(deptID, dto)

	return dto, nil
}

func (s *rawMaterialService) upsertSearch(deptID int, dto *model.RawMaterialDTO) {
	if dto == nil || dto.Name == nil {
		return
	}
	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "raw_material",
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

func (s *rawMaterialService) unlinkSearch(id int) {
	pubsub.PublishAsync("search:unlink", &searchmodel.UnlinkDoc{
		EntityType: "raw_material",
		EntityID:   int64(id),
	})
}

func (s *rawMaterialService) GetByID(ctx context.Context, deptID int, id int) (*model.RawMaterialDTO, error) {
	return cache.Get(kRawMaterialByID(deptID, id), cache.TTLMedium, func() (*model.RawMaterialDTO, error) {
		return s.repo.GetByID(ctx, deptID, id)
	})
}

func (s *rawMaterialService) List(ctx context.Context, deptID int, categoryID *int, q table.TableQuery) (table.TableListResult[model.RawMaterialDTO], error) {
	type boxed = table.TableListResult[model.RawMaterialDTO]
	key := kRawMaterialList(deptID, categoryID, q)

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

func (s *rawMaterialService) Delete(ctx context.Context, deptID int, id int) error {
	if err := s.repo.Delete(ctx, deptID, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kRawMaterialAll(deptID)...)
	cache.InvalidateKeys(kRawMaterialByID(deptID, id))

	s.unlinkSearch(id)
	return nil
}

func (s *rawMaterialService) Search(ctx context.Context, deptID int, categoryID *int, q dbutils.SearchQuery) (dbutils.SearchResult[model.RawMaterialDTO], error) {
	type boxed = dbutils.SearchResult[model.RawMaterialDTO]
	key := kRawMaterialSearch(deptID, categoryID, q)

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
