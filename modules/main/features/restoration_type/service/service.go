package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/restoration_type/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/module"
	searchmodel "github.com/khiemnd777/andy_api/shared/modules/search/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type RestorationTypeService interface {
	Create(ctx context.Context, deptID int, input model.RestorationTypeDTO) (*model.RestorationTypeDTO, error)
	Update(ctx context.Context, deptID int, input model.RestorationTypeDTO) (*model.RestorationTypeDTO, error)
	GetByID(ctx context.Context, deptID int, id int) (*model.RestorationTypeDTO, error)
	List(ctx context.Context, deptID int, categoryID *int, query table.TableQuery) (table.TableListResult[model.RestorationTypeDTO], error)
	Search(ctx context.Context, deptID int, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.RestorationTypeDTO], error)
	Delete(ctx context.Context, deptID int, id int) error
}

type restorationTypeService struct {
	repo repository.RestorationTypeRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewRestorationTypeService(repo repository.RestorationTypeRepository, deps *module.ModuleDeps[config.ModuleConfig]) RestorationTypeService {
	return &restorationTypeService{repo: repo, deps: deps}
}

func kRestorationTypeByID(id int) string {
	return fmt.Sprintf("restoration_type:id:%d", id)
}

func kRestorationTypeAll(deptID int) []string {
	return []string{
		kRestorationTypeListAll(deptID),
		kRestorationTypeSearchAll(deptID),
	}
}

func kRestorationTypeListAll(deptID int) string {
	return fmt.Sprintf("restoration_type:list:dpt%d:*", deptID)
}

func kRestorationTypeSearchAll(deptID int) string {
	return fmt.Sprintf("restoration_type:search:dpt%d:*", deptID)
}

func kRestorationTypeList(deptID int, categoryID *int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	cid := 0
	if categoryID != nil {
		cid = *categoryID
	}
	return fmt.Sprintf("restoration_type:list:dpt%d:c%d:l%d:p%d:o%s:d%s", deptID, cid, q.Limit, q.Page, orderBy, q.Direction)
}

func kRestorationTypeSearch(deptID int, categoryID *int, q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	cid := 0
	if categoryID != nil {
		cid = *categoryID
	}
	return fmt.Sprintf("restoration_type:search:dpt%d:c%d:k%s:l%d:p%d:o%s:d%s", deptID, cid, q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

func (s *restorationTypeService) Create(ctx context.Context, deptID int, input model.RestorationTypeDTO) (*model.RestorationTypeDTO, error) {
	dto, err := s.repo.Create(ctx, deptID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kRestorationTypeByID(dto.ID))
	}
	cache.InvalidateKeys(kRestorationTypeAll(deptID)...)

	s.upsertSearch(deptID, dto)

	return dto, nil
}

func (s *restorationTypeService) Update(ctx context.Context, deptID int, input model.RestorationTypeDTO) (*model.RestorationTypeDTO, error) {
	dto, err := s.repo.Update(ctx, deptID, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kRestorationTypeByID(dto.ID))
	}
	cache.InvalidateKeys(kRestorationTypeAll(deptID)...)

	s.upsertSearch(deptID, dto)

	return dto, nil
}

func (s *restorationTypeService) upsertSearch(deptID int, dto *model.RestorationTypeDTO) {
	if dto == nil || dto.Name == nil {
		return
	}
	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "restoration_type",
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

func (s *restorationTypeService) unlinkSearch(id int) {
	pubsub.PublishAsync("search:unlink", &searchmodel.UnlinkDoc{
		EntityType: "restoration_type",
		EntityID:   int64(id),
	})
}

func (s *restorationTypeService) GetByID(ctx context.Context, deptID, id int) (*model.RestorationTypeDTO, error) {
	return cache.Get(kRestorationTypeByID(id), cache.TTLMedium, func() (*model.RestorationTypeDTO, error) {
		return s.repo.GetByID(ctx, deptID, id)
	})
}

func (s *restorationTypeService) List(ctx context.Context, deptID int, categoryID *int, q table.TableQuery) (table.TableListResult[model.RestorationTypeDTO], error) {
	type boxed = table.TableListResult[model.RestorationTypeDTO]
	key := kRestorationTypeList(deptID, categoryID, q)

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

func (s *restorationTypeService) Delete(ctx context.Context, deptID int, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kRestorationTypeAll(deptID)...)
	cache.InvalidateKeys(kRestorationTypeByID(id))

	s.unlinkSearch(id)
	return nil
}

func (s *restorationTypeService) Search(ctx context.Context, deptID int, categoryID *int, q dbutils.SearchQuery) (dbutils.SearchResult[model.RestorationTypeDTO], error) {
	type boxed = dbutils.SearchResult[model.RestorationTypeDTO]
	key := kRestorationTypeSearch(deptID, categoryID, q)

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
