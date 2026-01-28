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
	GetByID(ctx context.Context, id int) (*model.RawMaterialDTO, error)
	List(ctx context.Context, categoryID *int, query table.TableQuery) (table.TableListResult[model.RawMaterialDTO], error)
	Search(ctx context.Context, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.RawMaterialDTO], error)
	Delete(ctx context.Context, id int) error
}

type rawMaterialService struct {
	repo repository.RawMaterialRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewRawMaterialService(repo repository.RawMaterialRepository, deps *module.ModuleDeps[config.ModuleConfig]) RawMaterialService {
	return &rawMaterialService{repo: repo, deps: deps}
}

func kRawMaterialByID(id int) string {
	return fmt.Sprintf("raw_material:id:%d", id)
}

func kRawMaterialAll() []string {
	return []string{
		kRawMaterialListAll(),
		kRawMaterialSearchAll(),
	}
}

func kRawMaterialListAll() string {
	return "raw_material:list:*"
}

func kRawMaterialSearchAll() string {
	return "raw_material:search:*"
}

func kRawMaterialList(categoryID *int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	cid := 0
	if categoryID != nil {
		cid = *categoryID
	}
	return fmt.Sprintf("raw_material:list:c%d:l%d:p%d:o%s:d%s", cid, q.Limit, q.Page, orderBy, q.Direction)
}

func kRawMaterialSearch(categoryID *int, q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	cid := 0
	if categoryID != nil {
		cid = *categoryID
	}
	return fmt.Sprintf("raw_material:search:c%d:k%s:l%d:p%d:o%s:d%s", cid, q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

func (s *rawMaterialService) Create(ctx context.Context, deptID int, input model.RawMaterialDTO) (*model.RawMaterialDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kRawMaterialByID(dto.ID))
	}
	cache.InvalidateKeys(kRawMaterialAll()...)

	s.upsertSearch(deptID, dto)

	return dto, nil
}

func (s *rawMaterialService) Update(ctx context.Context, deptID int, input model.RawMaterialDTO) (*model.RawMaterialDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kRawMaterialByID(dto.ID))
	}
	cache.InvalidateKeys(kRawMaterialAll()...)

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

func (s *rawMaterialService) GetByID(ctx context.Context, id int) (*model.RawMaterialDTO, error) {
	return cache.Get(kRawMaterialByID(id), cache.TTLMedium, func() (*model.RawMaterialDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

func (s *rawMaterialService) List(ctx context.Context, categoryID *int, q table.TableQuery) (table.TableListResult[model.RawMaterialDTO], error) {
	type boxed = table.TableListResult[model.RawMaterialDTO]
	key := kRawMaterialList(categoryID, q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.List(ctx, categoryID, q)
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

func (s *rawMaterialService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kRawMaterialAll()...)
	cache.InvalidateKeys(kRawMaterialByID(id))

	s.unlinkSearch(id)
	return nil
}

func (s *rawMaterialService) Search(ctx context.Context, categoryID *int, q dbutils.SearchQuery) (dbutils.SearchResult[model.RawMaterialDTO], error) {
	type boxed = dbutils.SearchResult[model.RawMaterialDTO]
	key := kRawMaterialSearch(categoryID, q)

	ptr, err := cache.Get(key, cache.TTLMedium, func() (*boxed, error) {
		res, e := s.repo.Search(ctx, categoryID, q)
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
