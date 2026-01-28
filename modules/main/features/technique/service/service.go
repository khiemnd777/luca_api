package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/modules/main/features/technique/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/module"
	searchmodel "github.com/khiemnd777/andy_api/shared/modules/search/model"
	"github.com/khiemnd777/andy_api/shared/pubsub"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/khiemnd777/andy_api/shared/utils/table"
)

type TechniqueService interface {
	Create(ctx context.Context, deptID int, input model.TechniqueDTO) (*model.TechniqueDTO, error)
	Update(ctx context.Context, deptID int, input model.TechniqueDTO) (*model.TechniqueDTO, error)
	GetByID(ctx context.Context, id int) (*model.TechniqueDTO, error)
	List(ctx context.Context, categoryID *int, query table.TableQuery) (table.TableListResult[model.TechniqueDTO], error)
	Search(ctx context.Context, categoryID *int, query dbutils.SearchQuery) (dbutils.SearchResult[model.TechniqueDTO], error)
	Delete(ctx context.Context, id int) error
}

type techniqueService struct {
	repo repository.TechniqueRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewTechniqueService(repo repository.TechniqueRepository, deps *module.ModuleDeps[config.ModuleConfig]) TechniqueService {
	return &techniqueService{repo: repo, deps: deps}
}

func kTechniqueByID(id int) string {
	return fmt.Sprintf("technique:id:%d", id)
}

func kTechniqueAll() []string {
	return []string{
		kTechniqueListAll(),
		kTechniqueSearchAll(),
	}
}

func kTechniqueListAll() string {
	return "technique:list:*"
}

func kTechniqueSearchAll() string {
	return "technique:search:*"
}

func kTechniqueList(categoryID *int, q table.TableQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	cid := 0
	if categoryID != nil {
		cid = *categoryID
	}
	return fmt.Sprintf("technique:list:c%d:l%d:p%d:o%s:d%s", cid, q.Limit, q.Page, orderBy, q.Direction)
}

func kTechniqueSearch(categoryID *int, q dbutils.SearchQuery) string {
	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}
	cid := 0
	if categoryID != nil {
		cid = *categoryID
	}
	return fmt.Sprintf("technique:search:c%d:k%s:l%d:p%d:o%s:d%s", cid, q.Keyword, q.Limit, q.Page, orderBy, q.Direction)
}

func (s *techniqueService) Create(ctx context.Context, deptID int, input model.TechniqueDTO) (*model.TechniqueDTO, error) {
	dto, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil && dto.ID > 0 {
		cache.InvalidateKeys(kTechniqueByID(dto.ID))
	}
	cache.InvalidateKeys(kTechniqueAll()...)

	s.upsertSearch(deptID, dto)

	return dto, nil
}

func (s *techniqueService) Update(ctx context.Context, deptID int, input model.TechniqueDTO) (*model.TechniqueDTO, error) {
	dto, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	if dto != nil {
		cache.InvalidateKeys(kTechniqueByID(dto.ID))
	}
	cache.InvalidateKeys(kTechniqueAll()...)

	s.upsertSearch(deptID, dto)

	return dto, nil
}

func (s *techniqueService) upsertSearch(deptID int, dto *model.TechniqueDTO) {
	if dto == nil || dto.Name == nil {
		return
	}
	pubsub.PublishAsync("search:upsert", &searchmodel.Doc{
		EntityType: "technique",
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

func (s *techniqueService) unlinkSearch(id int) {
	pubsub.PublishAsync("search:unlink", &searchmodel.UnlinkDoc{
		EntityType: "technique",
		EntityID:   int64(id),
	})
}

func (s *techniqueService) GetByID(ctx context.Context, id int) (*model.TechniqueDTO, error) {
	return cache.Get(kTechniqueByID(id), cache.TTLMedium, func() (*model.TechniqueDTO, error) {
		return s.repo.GetByID(ctx, id)
	})
}

func (s *techniqueService) List(ctx context.Context, categoryID *int, q table.TableQuery) (table.TableListResult[model.TechniqueDTO], error) {
	type boxed = table.TableListResult[model.TechniqueDTO]
	key := kTechniqueList(categoryID, q)

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

func (s *techniqueService) Delete(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	cache.InvalidateKeys(kTechniqueAll()...)
	cache.InvalidateKeys(kTechniqueByID(id))

	s.unlinkSearch(id)
	return nil
}

func (s *techniqueService) Search(ctx context.Context, categoryID *int, q dbutils.SearchQuery) (dbutils.SearchResult[model.TechniqueDTO], error) {
	type boxed = dbutils.SearchResult[model.TechniqueDTO]
	key := kTechniqueSearch(categoryID, q)

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
