package service

import (
	"context"
	"fmt"

	"github.com/khiemnd777/andy_api/modules/main/config"
	relation "github.com/khiemnd777/andy_api/modules/main/features/__relation/policy"
	"github.com/khiemnd777/andy_api/modules/main/features/__relation/repository"
	"github.com/khiemnd777/andy_api/shared/cache"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	dbutils "github.com/khiemnd777/andy_api/shared/db/utils"
	"github.com/khiemnd777/andy_api/shared/module"
	tableutils "github.com/khiemnd777/andy_api/shared/utils/table"
)

type RelationService struct {
	repo *repository.RelationRepository
	deps *module.ModuleDeps[config.ModuleConfig]
}

func NewRelationService(repo *repository.RelationRepository, deps *module.ModuleDeps[config.ModuleConfig]) *RelationService {
	return &RelationService{
		repo: repo,
		deps: deps,
	}
}

func (s *RelationService) List(
	ctx context.Context,
	key string,
	mainID int,
	q tableutils.TableQuery,
) (any, error) {

	cfg, err := relation.GetConfig(key)
	if err != nil {
		return nil, nil
	}

	if cfg.GetRefList == nil {
		return nil, nil
	}

	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}

	cKey := fmt.Sprintf(cfg.GetRefList.CachePrefix+":%s:%d:l%d:p%d:o%s:d%s", key, mainID, q.Limit, q.Page, orderBy, q.Direction)

	return cache.Get(cKey, cache.TTLShort, func() (*any, error) {
		tx, err := s.deps.Ent.(*generated.Client).Tx(ctx)
		if err != nil {
			return nil, fmt.Errorf("relation.List: cannot start tx: %w", err)
		}
		defer func() {
			_ = tx.Rollback()
		}()

		result, err := s.repo.List(ctx, tx, cfg, mainID, q)
		if err != nil {
			return nil, err
		}

		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("relation.List commit: %w", err)
		}

		return &result, nil
	})
}

func (s *RelationService) Search(ctx context.Context, key string, q dbutils.SearchQuery) (any, error) {
	cfg, err := relation.GetConfig(key)
	if err != nil {
		return nil, nil
	}

	if cfg.GetRefList == nil {
		return nil, nil
	}

	orderBy := ""
	if q.OrderBy != nil {
		orderBy = *q.OrderBy
	}

	cKey := fmt.Sprintf(cfg.GetRefList.CachePrefix+":%s:k%s:l%d:p%d:o%s:d%s", key, q.Keyword, q.Limit, q.Page, orderBy, q.Direction)

	return cache.Get(cKey, cache.TTLShort, func() (*any, error) {
		tx, err := s.deps.Ent.(*generated.Client).Tx(ctx)
		if err != nil {
			return nil, fmt.Errorf("relation.List: cannot start tx: %w", err)
		}
		defer func() {
			_ = tx.Rollback()
		}()

		result, err := s.repo.Search(ctx, tx, cfg, q)
		if err != nil {
			return nil, err
		}

		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("relation.List commit: %w", err)
		}

		return &result, nil
	})
}
